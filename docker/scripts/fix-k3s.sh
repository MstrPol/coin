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
chmod +x "${ROOT}/scripts/register-stack-endpoints.sh"
"${ROOT}/scripts/register-stack-endpoints.sh"

echo "k3s finalized"
