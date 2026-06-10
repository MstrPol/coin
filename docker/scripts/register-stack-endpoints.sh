#!/usr/bin/env bash
# Service+Endpoints в k3s для сервисов docker-compose (pods резолвят hostnames).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

container_ip() {
  local service="$1"
  docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
    "$(docker compose ps -q "${service}")" 2>/dev/null || true
}

apply_endpoints() {
  local name="$1"
  local ip="$2"
  shift 2
  # remaining args: portName:port pairs
  if [[ -z "${ip}" ]]; then
    echo "skip ${name}: container not running" >&2
    return 0
  fi

  local svc_ports="" ep_ports="" pair pname port
  for pair in "$@"; do
    pname="${pair%%:*}"
    port="${pair##*:}"
    svc_ports+="
    - name: ${pname}
      port: ${port}
      targetPort: ${port}"
    ep_ports+="
      - name: ${pname}
        port: ${port}"
  done

  echo "==> k3s Endpoints ${name} → ${ip}"
  docker compose exec -T k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: ${name}
  namespace: default
spec:
  ports:${svc_ports}
---
apiVersion: v1
kind: Endpoints
metadata:
  name: ${name}
  namespace: default
subsets:
  - addresses:
      - ip: ${ip}
    ports:${ep_ports}
EOF
}

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
COIN_API_PORT="${COIN_API_PORT:-8090}"

apply_endpoints jenkins "$(container_ip jenkins)" http:8080 agent:50000
apply_endpoints nexus "$(container_ip nexus)" ui:"${NEXUS_HTTP_PORT}" docker:"${NEXUS_DOCKER_PORT}"
apply_endpoints gitea "$(container_ip gitea)" http:3000
apply_endpoints coin-api "$(container_ip coin-api)" http:"${COIN_API_PORT}"

echo "stack Endpoints registered in k3s (default namespace)"
