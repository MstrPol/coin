#!/usr/bin/env bash
# Инициализация Jenkins dev-stand: ожидание UI, Endpoints в k3s.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
set -a && source "${ROOT}/.env" && set +a

JENKINS_HTTP_PORT="${JENKINS_HTTP_PORT:-8080}"
JENKINS_ADMIN_USER="${JENKINS_ADMIN_USER:-admin}"
JENKINS_ADMIN_PASSWORD="${JENKINS_ADMIN_PASSWORD:-admin}"

echo "==> waiting for Jenkins"
for _ in $(seq 1 90); do
  if curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" \
    "http://localhost:${JENKINS_HTTP_PORT}/login"; then
    break
  fi
  sleep 3
done

JENKINS_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q jenkins)" 2>/dev/null || true)"
if [[ -z "${JENKINS_IP}" ]]; then
  echo "failed to resolve jenkins container IP" >&2
  exit 1
fi

echo "==> registering jenkins in k3s (Endpoints ${JENKINS_IP}:8080, :50000)"
docker compose exec -T k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: jenkins
  namespace: default
spec:
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: agent
      port: 50000
      targetPort: 50000
---
apiVersion: v1
kind: Endpoints
metadata:
  name: jenkins
  namespace: default
subsets:
  - addresses:
      - ip: ${JENKINS_IP}
    ports:
      - name: http
        port: 8080
      - name: agent
        port: 50000
EOF

echo "Jenkins ready: http://localhost:${JENKINS_HTTP_PORT} (${JENKINS_ADMIN_USER} / ${JENKINS_ADMIN_PASSWORD})"
