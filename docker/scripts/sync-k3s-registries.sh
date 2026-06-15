#!/usr/bin/env bash
# containerd auth для pull из Nexus (pods: nexus:8082/coin-docker/...).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"
REGISTRY="nexus:${NEXUS_DOCKER_PORT}"
OUT="${ROOT}/k3s/registries.yaml"

cat > "${OUT}" <<EOF
mirrors:
  "${REGISTRY}":
    endpoint:
      - "http://${REGISTRY}"

configs:
  "${REGISTRY}":
    auth:
      username: ${NEXUS_DOCKER_USER}
      password: ${NEXUS_DOCKER_PASSWORD}
    tls:
      insecure_skip_verify: true
EOF

echo "==> k3s registries.yaml → ${REGISTRY} (user ${NEXUS_DOCKER_USER})"

if docker compose ps -q k3s >/dev/null 2>&1; then
  echo "==> restart k3s (reload registries)"
  docker compose restart k3s >/dev/null
  for _ in $(seq 1 60); do
    if docker compose exec -T k3s kubectl get nodes >/dev/null 2>&1; then
      break
    fi
    sleep 2
  done
fi
