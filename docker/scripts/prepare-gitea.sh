#!/usr/bin/env bash
# Инициализация Gitea: admin, репозитории, Service+Endpoints в k3s.
# migrate выполняет entrypoint контейнera → конфиг в /data/gitea/conf/app.ini (writable).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
GITEA_EMAIL="${GITEA_EMAIL:-coin@local}"
GITEA_CONFIG="/data/gitea/conf/app.ini"

echo "==> waiting for Gitea (entrypoint runs migrate into ${GITEA_CONFIG})"
for i in $(seq 1 90); do
  if docker compose exec -u git -T gitea test -f "${GITEA_CONFIG}" 2>/dev/null; then
    break
  fi
  sleep 2
done

for i in $(seq 1 60); do
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

create_repo() {
  local name="$1"
  curl -sf -u "${GITEA_USER}:${GITEA_PASSWORD}" \
    -X POST "http://localhost:${GITEA_HTTP_PORT}/api/v1/user/repos" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"${name}\",\"private\":false,\"default_branch\":\"main\"}" \
    >/dev/null 2>&1 || true
}

echo "==> creating repositories"
create_repo "demo-python-uv"
create_repo "coin"

echo "==> registering gitea in k3s (Endpoints for pod checkout)"
GITEA_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q gitea)")"
if [[ -z "${GITEA_IP}" ]]; then
  echo "failed to resolve gitea container IP" >&2
  exit 1
fi

docker compose exec -T k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: gitea
  namespace: default
spec:
  ports:
    - port: 3000
      targetPort: 3000
---
apiVersion: v1
kind: Endpoints
metadata:
  name: gitea
  namespace: default
subsets:
  - addresses:
      - ip: ${GITEA_IP}
    ports:
      - port: 3000
EOF

REGISTRY_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q registry)" 2>/dev/null || true)"
if [[ -n "${REGISTRY_IP}" ]]; then
  echo "==> registering registry in k3s (Endpoints ${REGISTRY_IP}:5000)"
  docker compose exec -T k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: registry
  namespace: default
spec:
  ports:
    - port: 5000
      targetPort: 5000
---
apiVersion: v1
kind: Endpoints
metadata:
  name: registry
  namespace: default
subsets:
  - addresses:
      - ip: ${REGISTRY_IP}
    ports:
      - port: 5000
EOF
fi

echo "Gitea ready: http://localhost:${GITEA_HTTP_PORT} (${GITEA_USER} / ${GITEA_PASSWORD})"
