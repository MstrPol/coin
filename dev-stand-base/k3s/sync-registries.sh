#!/usr/bin/env bash
# containerd auth для pull из Nexus (pods: nexus:8082/coin-docker/...).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# shellcheck disable=SC1091
set -a && source "${ROOT}/.env" && set +a

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
