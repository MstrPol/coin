#!/usr/bin/env bash
# Инициализация k3s dev-stand: API, CoreDNS, registries, Endpoints для gitea/nexus.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
set -a && source "${ROOT}/.env" && set +a

K3S_DASHBOARD_PORT="${K3S_DASHBOARD_PORT:-8443}"
NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"

container_ip() {
  local service="$1"
  local id
  id="$(docker compose ps -q "${service}" 2>/dev/null || true)"
  if [[ -z "${id}" && "${service}" == "jenkins" ]]; then
    # orphan из старого compose-проекта (dev-stand_*)
    id="$(docker ps -q -f name=jenkins 2>/dev/null | head -n1 || true)"
  fi
  if [[ -z "${id}" ]]; then
    return 0
  fi
  docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${id}" 2>/dev/null || true
}

apply_endpoints() {
  local name="$1"
  local ip="$2"
  shift 2
  local pair pname port svc_ports="" ep_ports=""

  if [[ -z "${ip}" ]]; then
    echo "skip ${name}: container not running" >&2
    return 0
  fi

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

echo "==> waiting for k3s kubeconfig"
for _ in $(seq 1 60); do
  if docker compose exec -T k3s test -f /output/kubeconfig.yaml 2>/dev/null; then
    break
  fi
  sleep 2
done

echo "==> waiting for k3s API"
for _ in $(seq 1 60); do
  if docker compose exec -T k3s kubectl get nodes >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

docker compose exec -T k3s sh -c '
  cp /output/kubeconfig.yaml /output/config
  sed -i "s|127.0.0.1|k3s|g" /output/config
  sed -i "s|localhost|k3s|g" /output/config
'

echo "==> removing NotReady nodes"
docker compose exec -T k3s kubectl get nodes -o json \
  | python3 -c '
import json, sys
data = json.load(sys.stdin)
for node in data.get("items", []):
    name = node["metadata"]["name"]
    ready = next((c["status"] for c in node.get("status", {}).get("conditions", []) if c["type"] == "Ready"), "Unknown")
    if ready != "True":
        print(name)
' | while read -r node; do
  [[ -n "${node}" ]] || continue
  echo "  delete node ${node}"
  # NotReady node: pods на нём не завершаются без force — иначе kubectl hang.
  docker compose exec -T k3s kubectl delete node "${node}" --ignore-not-found --wait=false
done

echo "==> restarting CoreDNS"
# --wait=false: не ждать Terminating на старых NotReady nodes (иначе hang forever).
docker compose exec -T k3s kubectl delete pod -n kube-system -l k8s-app=kube-dns \
  --ignore-not-found --wait=false --force --grace-period=0
for _ in $(seq 1 30); do
  phase="$(docker compose exec -T k3s kubectl get pods -n kube-system -l k8s-app=kube-dns \
    --field-selector=status.phase=Running \
    -o jsonpath='{.items[0].status.phase}' 2>/dev/null || true)"
  if [[ "${phase}" == "Running" ]]; then
    break
  fi
  sleep 2
done

"${ROOT}/k3s/sync-registries.sh"

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

apply_endpoints nexus "$(container_ip nexus)" ui:"${NEXUS_HTTP_PORT}" docker:"${NEXUS_DOCKER_PORT}"
apply_endpoints gitea "$(container_ip gitea)" http:3000
apply_endpoints jenkins "$(container_ip jenkins)" http:8080 agent:50000

echo "k3s ready: API https://localhost:${K3S_API_PORT:-6443}, Dashboard https://localhost:${K3S_DASHBOARD_PORT} (make dashboard-up)"
