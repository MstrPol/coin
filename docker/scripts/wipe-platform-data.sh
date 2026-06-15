#!/usr/bin/env bash
# Полная очистка данных coin-api (GP, components, projects) без пересоздания postgres volume.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

echo "==> wiping platform data in postgres"
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
DELETE FROM component_compatibility;
DELETE FROM component_versions;
DELETE FROM components;
DELETE FROM gp_profiles;
COMMIT;
SQL

echo "platform data wiped (остаётся только platform_settings)"
