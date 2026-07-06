#!/usr/bin/env bash
# Local pilot: publish lib (Nexus) + branching-model + GP profile/release (2-pin + embedded pipeline).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_PUBLISHER_API_KEY:-dev-local-admin-key}"
AUTH=(-H "X-API-Key: ${KEY}" -H "Content-Type: application/json")

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
need curl
need jq

api_post() {
  local path="$1" body="$2"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X POST "${API}${path}" "${AUTH[@]}" -d "${body}")"
  if [[ "${code}" != "201" && "${code}" != "200" && "${code}" != "409" ]]; then
    echo "POST ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  rm -f "${tmp}"
}

component_version() {
  local typ="$1" name="$2"
  curl -fsS "${API}/v1/admin/components/${typ}/${name}/versions" "${AUTH[@]}" 2>/dev/null \
    | jq -r '
        [.items[] | select(.status=="published") | .version]
        | map(select(test("^[0-9]+\\.[0-9]+\\.[0-9]+$")))
        | sort_by(split(".") | map(tonumber))
        | last // empty
      '
}

echo "==> cleanup legacy pipeline-bundle components"
docker compose -f "${ROOT}/compose.yml" exec -T postgres psql -U "${POSTGRES_USER:-coin}" -d "${POSTGRES_DB:-coin}" <<'SQL' || true
DELETE FROM component_artifact_bodies
WHERE component_version_id IN (
  SELECT cv.id FROM component_versions cv
  JOIN components c ON c.id = cv.component_id
  WHERE c.type = 'pipeline-bundle'
);
DELETE FROM component_versions
WHERE component_id IN (SELECT id FROM components WHERE type = 'pipeline-bundle');
DELETE FROM components WHERE type = 'pipeline-bundle';
SQL

echo "==> cleanup legacy lib/platform-starter"
docker compose -f "${ROOT}/compose.yml" exec -T postgres psql -U "${POSTGRES_USER:-coin}" -d "${POSTGRES_DB:-coin}" <<'SQL' || true
DELETE FROM component_artifact_bodies
WHERE component_version_id IN (
  SELECT cv.id FROM component_versions cv
  JOIN components c ON c.id = cv.component_id
  WHERE c.type = 'lib' AND c.name = 'platform-starter'
);
DELETE FROM component_versions
WHERE component_id IN (SELECT id FROM components WHERE type = 'lib' AND name = 'platform-starter');
DELETE FROM components WHERE type = 'lib' AND name = 'platform-starter';
SQL

echo "==> publish lib/coin-lib@1.0.0"
chmod +x "${REPO_ROOT}/coin-lib/scripts/"*.sh "${REPO_ROOT}/coin-lib/scripts/lib/"*.sh
NEXUS_URL="${NEXUS_URL:-http://localhost:8081}" \
NEXUS_USER="${NEXUS_USER:-admin}" \
NEXUS_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}" \
  "${REPO_ROOT}/coin-lib/scripts/publish-lib.sh" 1.0.0

echo "==> publish branching-model/trunk-based@1.0.0"
chmod +x "${REPO_ROOT}/coin-branching-models/scripts/"*.sh "${REPO_ROOT}/coin-branching-models/scripts/lib/"*.sh
COIN_API_URL="${API}" \
COIN_API_KEY="${KEY}" \
  "${REPO_ROOT}/coin-branching-models/scripts/publish-branching-model.sh" trunk-based 1.0.0

AGENT_VER="$(component_version agent coin-agent)"
BRANCHING_VER="$(component_version branching-model trunk-based)"

for pair in "agent:${AGENT_VER}" "branching-model:${BRANCHING_VER}"; do
  if [[ -z "${pair#*:}" ]]; then
    echo "missing component version for ${pair%%:*}" >&2
    exit 1
  fi
done

GP_STACKS=(go-app go-app-docker)
GP_VER="${COIN_E2E_VERSION:-1.0.0}"
BRANCHING_MODEL="trunk-based"
DESTINATIONS="$(jq -n \
  --arg image "${COIN_GP_IMAGE_REGISTRY_PREFIX:-localhost:8082/coin-docker}" \
  --arg artifact "${COIN_GP_ARTIFACT_REPOSITORY_BASE:-http://nexus:8081/repository/maven-releases}" \
  '{imageRegistryPrefix: $image, buildCacheEnabled: true, artifactRepositoryBase: $artifact}')"

for stack in "${GP_STACKS[@]}"; do
  echo "==> create GP profile ${stack} (if missing)"
  api_post "/v1/admin/golden-paths/profiles" "$(jq -n --arg n "${stack}" '{name: $n, actor: "seed"}')" || true

  echo "==> publish GP ${stack}@${GP_VER} (agent@${AGENT_VER}, branching-model@${BRANCHING_VER}, embedded pipeline)"
  composition="$(jq -n \
    --arg agent "${AGENT_VER}" \
    --arg branching "${BRANCHING_VER}" \
    '{"agent": $agent, "branching-model": $branching}')"
  gp_body="$(jq -n \
    --arg ver "${GP_VER}" \
    --arg stack "${stack}" \
    --arg bm "${BRANCHING_MODEL}" \
    --arg agent "coin-agent" \
    --argjson comp "${composition}" \
    --argjson destinations "${DESTINATIONS}" \
    '{version: $ver, destinations: $destinations, agentStackName: $agent, composition: $comp, branchingModelName: $bm, actor: "seed"}')"
  gp_tmp="$(mktemp)"
  gp_code="$(curl -sS -o "${gp_tmp}" -w '%{http_code}' -X POST "${API}/v1/admin/golden-paths/${stack}/versions" "${AUTH[@]}" -d "${gp_body}")"
  if [[ "${gp_code}" != "201" && "${gp_code}" != "409" ]]; then
    echo "publish GP ${stack} failed HTTP ${gp_code}: $(cat "${gp_tmp}")" >&2
    rm -f "${gp_tmp}"
    exit 1
  fi
  rm -f "${gp_tmp}"

  echo "==> verify /golden-paths/${stack}/versions/${GP_VER}/manifest (branching + pipeline)"
  curl -fsS "${API}/v1/golden-paths/${stack}/versions/${GP_VER}/manifest" \
    -H "Authorization: Bearer ${COIN_API_TOKEN:-dev-local-token}" \
    | jq -e --arg bm "${BRANCHING_VER}" \
      '.branching.name == "trunk-based" and .branching.version == $bm' >/dev/null
  curl -fsS "${API}/v1/golden-paths/${stack}/versions/${GP_VER}/manifest" \
    -H "Authorization: Bearer ${COIN_API_TOKEN:-dev-local-token}" \
    | jq -e 'has("lib") | not' >/dev/null
  curl -fsS "${API}/v1/golden-paths/${stack}/versions/${GP_VER}/manifest" \
    -H "Authorization: Bearer ${COIN_API_TOKEN:-dev-local-token}" \
    | jq -e '.destinations.imageRegistryPrefix and .destinations.artifactRepositoryBase and (.pipeline.stages | length > 0) and (has("credentials") | not) and ((has("build") | not) or (.build | length == 0))' >/dev/null
done

echo "OK: jenkins-lib stack seeded (${GP_STACKS[*]}@${GP_VER}, trunk-based@${BRANCHING_VER}, coin-agent@${AGENT_VER})"

echo "==> Jenkins: coin-lib Nexus HTTP retriever (primary path)"
chmod +x "${ROOT}/scripts/coin-lib-http.sh"
"${ROOT}/scripts/coin-lib-http.sh"
