#!/usr/bin/env bash
# Финализация k3s: ghost-ноды, CoreDNS, Endpoints (вызывается из bootstrap).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

echo "==> removing NotReady nodes"
docker compose exec -T k3s kubectl get nodes -o json \
  | python3 -c '
import json, sys, subprocess
data = json.load(sys.stdin)
for node in data.get("items", []):
    name = node["metadata"]["name"]
    ready = next((c["status"] for c in node.get("status", {}).get("conditions", []) if c["type"] == "Ready"), "Unknown")
    if ready != "True":
        print(name)
' | while read -r node; do
  [[ -n "${node}" ]] || continue
  echo "  delete node ${node}"
  docker compose exec -T k3s kubectl delete node "${node}" --ignore-not-found
done

echo "==> restarting CoreDNS"
docker compose exec -T k3s kubectl delete pod -n kube-system -l k8s-app=kube-dns --ignore-not-found
for i in $(seq 1 30); do
  if docker compose exec -T k3s kubectl get pod -n kube-system -l k8s-app=kube-dns -o jsonpath='{.items[0].status.phase}' 2>/dev/null | grep -q Running; then
    break
  fi
  sleep 2
done

"${ROOT}/scripts/setup-jenkins-k8s-auth.sh"
"${ROOT}/scripts/register-jenkins-k8s-endpoints.sh"

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"

NEXUS_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q nexus)" 2>/dev/null || true)"
if [[ -n "${NEXUS_IP}" ]]; then
  echo "==> updating nexus Endpoints (${NEXUS_IP}:${NEXUS_HTTP_PORT}, :${NEXUS_DOCKER_PORT})"
  docker compose exec -T k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: nexus
  namespace: default
spec:
  ports:
    - name: ui
      port: ${NEXUS_HTTP_PORT}
      targetPort: ${NEXUS_HTTP_PORT}
    - name: docker
      port: ${NEXUS_DOCKER_PORT}
      targetPort: ${NEXUS_DOCKER_PORT}
---
apiVersion: v1
kind: Endpoints
metadata:
  name: nexus
  namespace: default
subsets:
  - addresses:
      - ip: ${NEXUS_IP}
    ports:
      - name: ui
        port: ${NEXUS_HTTP_PORT}
      - name: docker
        port: ${NEXUS_DOCKER_PORT}
EOF
fi

GITEA_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q gitea)" 2>/dev/null || true)"
if [[ -n "${GITEA_IP}" ]]; then
  echo "==> updating gitea Endpoints (${GITEA_IP}:3000)"
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
fi

echo "k3s finalized"
