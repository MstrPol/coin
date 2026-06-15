#!/usr/bin/env bash
# Local pilot: publish lib + gp-content + GP profile/release for go-app (5-slot model).
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
    | jq -r '[.items[] | select(.status=="published") | .version] | last // empty'
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
COIN_API_URL="${API}" \
COIN_API_KEY="${KEY}" \
  "${REPO_ROOT}/coin-lib/scripts/publish-lib.sh" 1.0.0

echo "==> publish gp-content/go-app@1.0.0"
chmod +x "${REPO_ROOT}/coin-gp-content/scripts/"*.sh "${REPO_ROOT}/coin-gp-content/scripts/lib/"*.sh
NEXUS_URL="${NEXUS_URL:-http://localhost:8081}" \
NEXUS_USER="${NEXUS_USER:-admin}" \
NEXUS_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}" \
COIN_API_URL="${API}" \
COIN_API_KEY="${KEY}" \
  "${REPO_ROOT}/coin-gp-content/scripts/publish-content.sh" go-app 1.0.0

JNLP_VER="$(component_version agent jnlp)"
GO_VER="$(component_version agent go)"
EXEC_VER="$(component_version executor coin-executor)"
LIB_VER="$(component_version lib coin-lib)"
CONTENT_VER="$(component_version gp-content go-app)"

for pair in "jnlp:${JNLP_VER}" "go:${GO_VER}" "executor:${EXEC_VER}" "lib:${LIB_VER}" "content:${CONTENT_VER}"; do
  if [[ -z "${pair#*:}" ]]; then
    echo "missing component version for ${pair%%:*}" >&2
    exit 1
  fi
done

echo "==> create GP profile go-app (if missing)"
api_post "/v1/admin/golden-paths/profiles" '{"name":"go-app","agentStack":"go","actor":"seed"}' || true

GP_VER="${COIN_E2E_VERSION:-1.0.1}"

echo "==> publish GP go-app@${GP_VER}"
composition="$(jq -n \
  --arg jnlp "${JNLP_VER}" \
  --arg agent "${GO_VER}" \
  --arg exec "${EXEC_VER}" \
  --arg lib "${LIB_VER}" \
  --arg content "${CONTENT_VER}" \
  '{jnlp: $jnlp, agent: $agent, executor: $exec, lib: $lib, "gp-content": $content}')"
gp_body="$(jq -n \
  --arg ver "${GP_VER}" \
  --argjson comp "${composition}" \
  '{version: $ver, composition: $comp, actor: "seed"}')"
gp_tmp="$(mktemp)"
gp_code="$(curl -sS -o "${gp_tmp}" -w '%{http_code}' -X POST "${API}/v1/admin/golden-paths/go-app/versions" "${AUTH[@]}" -d "${gp_body}")"
if [[ "${gp_code}" != "201" && "${gp_code}" != "409" ]]; then
  echo "publish GP failed HTTP ${gp_code}: $(cat "${gp_tmp}")" >&2
  rm -f "${gp_tmp}"
  exit 1
fi
rm -f "${gp_tmp}"

pin_enc="$(python3 -c "import urllib.parse; print(urllib.parse.quote('=${GP_VER}', safe=''))")"
echo "==> verify /golden-paths/go-app/version"
curl -fsS "${API}/v1/golden-paths/go-app/version?pin=${pin_enc}" \
  -H "Authorization: Bearer ${COIN_API_TOKEN:-dev-local-token}" \
  | jq -e '.library.version == "1.0.0"'

echo "OK: jenkins-lib stack seeded (go-app@${GP_VER})"
