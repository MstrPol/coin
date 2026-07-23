#!/usr/bin/env bash
# SA + bearer token для Jenkins → k3s API.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

echo "==> waiting for k3s API"
for _ in $(seq 1 60); do
  if docker compose exec -T k3s kubectl get nodes >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> applying jenkins-k8s ServiceAccount"
docker compose exec -T k3s kubectl apply -f /k3s/jenkins-sa.yaml

echo "==> writing jenkins API token to k3s-kubeconfig volume"
docker compose exec -T k3s sh -c '
  kubectl create token jenkins-k8s -n default --duration=8760h > /output/jenkins-token
  chmod 644 /output/jenkins-token
'

echo "jenkins-k8s token ready at /kubeconfig/jenkins-token (inside Jenkins container)"
