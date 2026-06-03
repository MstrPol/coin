#!/usr/bin/env bash
# Kubernetes Dashboard (v2.7) — браузерный UI для k3s.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

K3S_DASHBOARD_PORT="${K3S_DASHBOARD_PORT:-8443}"

echo "==> waiting for k3s API"
for i in $(seq 1 60); do
  if docker compose exec -T k3s kubectl get nodes >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> applying Kubernetes Dashboard manifests"
docker compose exec -T k3s kubectl apply -f /dashboard/recommended.yaml
docker compose exec -T k3s kubectl apply -f /dashboard/admin-user.yaml

echo "==> waiting for dashboard pod"
for i in $(seq 1 60); do
  ready="$(docker compose exec -T k3s kubectl -n kubernetes-dashboard get deploy/kubernetes-dashboard \
    -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
  if [[ "${ready}" == "1" ]]; then
    break
  fi
  sleep 2
done

TOKEN="$(docker compose exec -T k3s kubectl -n kubernetes-dashboard create token admin-user --duration=8760h 2>/dev/null | tr -d '\r\n')"

cat <<EOF

Kubernetes Dashboard: https://localhost:${K3S_DASHBOARD_PORT}
  (самоподписанный сертификат — подтвердите исключение в браузере)

Вход: Token
Токен (8760h): ${TOKEN}

Повторно получить токен: make -C docker dashboard-token

EOF
