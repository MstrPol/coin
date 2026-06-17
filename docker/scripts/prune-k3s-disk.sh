#!/usr/bin/env bash
# Освободить ephemeral-storage: Docker build cache (host) + k3s pods/images.
# k3s node disk-pressure taint снимается kubelet'ом после освобождения места;
# если taint залип — перезапуск k3s (см. COIN_RESTART_K3S_ON_DISK_PRESSURE).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

PRUNE_DOCKER="${COIN_PRUNE_DOCKER:-0}"
RESTART_K3S="${COIN_RESTART_K3S_ON_DISK_PRESSURE:-1}"

for arg in "$@"; do
  case "${arg}" in
    --docker) PRUNE_DOCKER=1 ;;
    --all) PRUNE_DOCKER=1 ;;
  esac
done

if [[ "${PRUNE_DOCKER}" == "1" ]]; then
  echo "==> prune host: Docker build cache + dangling images"
  docker builder prune -af >/dev/null 2>&1 || true
  docker image prune -af >/dev/null 2>&1 || true
fi

if ! docker compose ps -q k3s >/dev/null 2>&1; then
  echo "k3s not running — skip k3s prune"
  exit 0
fi

echo "==> prune k3s: stale Jenkins agent pods"
docker compose exec -T k3s sh -c '
  kubectl delete pods -n default -l jenkins=slave --field-selector=status.phase=Failed --ignore-not-found=true 2>/dev/null || true
  kubectl delete pods -n default -l jenkins=slave --field-selector=status.phase=Succeeded --ignore-not-found=true 2>/dev/null || true
  kubectl delete pods -n default -l jenkins=slave --field-selector=status.phase=Unknown --ignore-not-found=true 2>/dev/null || true
  kubectl delete pods -n default -l jenkins=slave --field-selector=status.phase=Pending --ignore-not-found=true 2>/dev/null || true
' || true

echo "==> prune k3s: unused container images (crictl)"
docker compose exec -T k3s sh -c 'crictl rmi --prune 2>/dev/null || true' || true

if [[ "${RESTART_K3S}" == "1" ]]; then
  taint="$(docker compose exec -T k3s kubectl get node -o jsonpath='{.items[0].spec.taints}' 2>/dev/null || true)"
  if [[ "${taint}" == *disk-pressure* ]]; then
    echo "==> k3s node still has disk-pressure taint — restarting k3s"
    docker compose restart k3s
    for _ in $(seq 1 30); do
      if docker compose exec -T k3s kubectl get nodes >/dev/null 2>&1; then
        break
      fi
      sleep 2
    done
    docker compose exec -T k3s kubectl wait --for=condition=Ready node --all --timeout=60s >/dev/null 2>&1 || true
    bash "${ROOT}/scripts/register-stack-endpoints.sh" >/dev/null 2>&1 || true
  fi
fi

echo "==> k3s disk prune done"
