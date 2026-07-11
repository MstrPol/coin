#!/usr/bin/env bash
# Чистый лист для E2E: postgres (coin-api), все артефакты Nexus, Jenkins jobs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_ADMIN_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}"
NEXUS_MAVEN_RELEASES="${NEXUS_MAVEN_RELEASES:-maven-releases}"
NEXUS_MAVEN_SNAPSHOTS="${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
auth="admin:${NEXUS_ADMIN_PASSWORD}"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
need curl
need jq
need docker

echo "==> [1/3] wipe coin-api postgres"
docker compose exec -T postgres psql -U "${POSTGRES_USER:-coin}" -d "${POSTGRES_DB:-coin}" <<'SQL'
BEGIN;
DELETE FROM gp_artifact_bodies;
DELETE FROM gp_composition;
DELETE FROM gp_releases;
DELETE FROM catalog_policy;
DELETE FROM canary_policy;
DELETE FROM build_reports;
DELETE FROM projects;
DELETE FROM audit_log;
DELETE FROM component_artifact_bodies;
DELETE FROM component_compatibility;
DELETE FROM component_versions;
DELETE FROM components;
DELETE FROM gp_profiles;
COMMIT;
SQL

nexus_purge_repo() {
  local repo="$1"
  local continuation=""
  local deleted=0
  while true; do
    local url="http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/components?repository=${repo}"
    [[ -n "${continuation}" ]] && url="${url}&continuationToken=${continuation}"
    local resp
    resp="$(curl -sf -u "${auth}" "${url}")"
    while IFS= read -r id; do
      [[ -z "${id}" ]] && continue
      curl -sf -u "${auth}" -X DELETE \
        "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/components/${id}" >/dev/null
      deleted=$((deleted + 1))
    done < <(python3 -c 'import json,sys; d=json.load(sys.stdin); print("\n".join(i["id"] for i in d.get("items",[])))' <<<"${resp}")
    continuation="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("continuationToken") or "")' <<<"${resp}")"
    [[ -z "${continuation}" ]] && break
  done
  echo "    ${repo}: deleted ${deleted} component(s)"
}

echo "==> [2/3] purge Nexus repositories"
for repo in "${NEXUS_MAVEN_RELEASES}" "${NEXUS_MAVEN_SNAPSHOTS}" "${NEXUS_DOCKER_REPO}" "${NEXUS_DOCKER_CACHE_REPO:-coin-cache}"; do
  nexus_purge_repo "${repo}" || echo "    ${repo}: skip (repo missing or empty)"
done

echo "==> [3/3] wipe Jenkins jobs + platform CASC (keep 00-base.yaml)"
JENKINS_USER="${JENKINS_ADMIN_USER:-admin}"
JENKINS_PASS="${JENKINS_ADMIN_PASSWORD:-admin}"
JENKINS_PORT="${JENKINS_HTTP_PORT:-8080}"

if docker compose ps -q jenkins >/dev/null 2>&1; then
  docker compose exec -T -u root jenkins sh -c '
    rm -rf /var/jenkins_home/jobs/*
    find /var/jenkins_home/casc-config -type f -name "*.yaml" ! -name "00-base.yaml" -delete
  ' || true

  # Сброс in-memory WorkflowRun: wipe jobs/* на живом Jenkins оставляет зомби-сборки
  # (UI: building, log/build.xml отсутствуют). Restart обязателен.
  echo "    restarting Jenkins to drop in-memory job state..."
  docker compose restart jenkins
  for _ in $(seq 1 60); do
    if curl -sf -o /dev/null -u "${JENKINS_USER}:${JENKINS_PASS}" \
      "http://localhost:${JENKINS_PORT}/login"; then
      break
    fi
    sleep 3
  done
  echo "    Jenkins jobs cleared"
else
  echo "    Jenkins not running — skip job wipe"
fi

bash "${ROOT}/scripts/prune-k3s-disk.sh" || true

echo "==> done: clean slate for E2E"
