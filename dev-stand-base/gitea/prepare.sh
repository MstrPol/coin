#!/usr/bin/env bash
# Инициализация Gitea dev-stand: ожидание migrate и создание admin.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
set -a && source "${ROOT}/.env" && set +a

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_CONFIG="/data/gitea/conf/app.ini"

echo "==> waiting for Gitea (entrypoint runs migrate into ${GITEA_CONFIG})"
for _ in $(seq 1 90); do
  if docker compose exec -u git -T gitea test -f "${GITEA_CONFIG}" 2>/dev/null; then
    break
  fi
  sleep 2
done

for _ in $(seq 1 60); do
  if curl -sf "http://localhost:${GITEA_HTTP_PORT}/api/v1/version" >/dev/null 2>&1; then
    break
  fi
  if curl -sf "http://localhost:${GITEA_HTTP_PORT}/" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! docker compose exec -u git -T gitea test -f "${GITEA_CONFIG}"; then
  echo "Gitea config not found at ${GITEA_CONFIG}" >&2
  docker compose logs --tail=50 gitea >&2 || true
  exit 1
fi

echo "==> gitea admin user"
docker compose exec -u git -T gitea gitea admin user create \
  --admin \
  --username "${GITEA_USER}" \
  --password "${GITEA_PASSWORD}" \
  --email "${GITEA_EMAIL}" \
  --config "${GITEA_CONFIG}" 2>/dev/null \
  || docker compose exec -u git -T gitea gitea admin user change-password \
       --username "${GITEA_USER}" \
       --password "${GITEA_PASSWORD}" \
       --config "${GITEA_CONFIG}" 2>/dev/null \
  || true

echo "Gitea ready: http://localhost:${GITEA_HTTP_PORT} (${GITEA_USER} / ${GITEA_PASSWORD})"
