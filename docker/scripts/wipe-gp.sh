#!/usr/bin/env bash
# Удалить все GP из coin-api и manifest blob/pointer в Nexus (coin/manifest/*).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_ADMIN_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}"
NEXUS_MAVEN_RELEASES="${NEXUS_MAVEN_RELEASES:-maven-releases}"
NEXUS_MAVEN_SNAPSHOTS="${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}"
auth="admin:${NEXUS_ADMIN_PASSWORD}"

echo "==> deleting GP data in postgres"
docker compose exec -T postgres psql -U "${POSTGRES_USER:-coin}" -d "${POSTGRES_DB:-coin}" <<'SQL'
BEGIN;
DELETE FROM gp_artifact_bodies;
DELETE FROM gp_composition;
DELETE FROM gp_releases;
DELETE FROM catalog_policy;
DELETE FROM canary_policy;
DELETE FROM gp_profiles;
DELETE FROM audit_log WHERE entity_type IN ('gp_release', 'gp_profile');
COMMIT;
SQL

nexus_delete_manifest_assets() {
  local repo="$1"
  local continuation=""
  local deleted=0
  while true; do
    local url="http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/search/assets?repository=${repo}&group=coin.manifest"
    [[ -n "${continuation}" ]] && url="${url}&continuationToken=${continuation}"
    local resp
    resp="$(curl -sf -u "${auth}" "${url}")"
    while IFS= read -r id; do
      [[ -z "${id}" ]] && continue
      curl -sf -u "${auth}" -X DELETE \
        "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/assets/${id}" >/dev/null
      deleted=$((deleted + 1))
    done < <(python3 -c 'import json,sys; d=json.load(sys.stdin); print("\n".join(i["id"] for i in d.get("items",[])))' <<<"${resp}")
    continuation="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("continuationToken") or "")' <<<"${resp}")"
    [[ -z "${continuation}" ]] && break
  done
  echo "    ${repo}: deleted ${deleted} manifest asset(s)"
}

echo "==> purging Nexus manifest assets (coin.manifest)"
nexus_delete_manifest_assets "${NEXUS_MAVEN_RELEASES}"
nexus_delete_manifest_assets "${NEXUS_MAVEN_SNAPSHOTS}"

echo "==> done: GP registry empty, Nexus coin.manifest cleared"
