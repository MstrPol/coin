#!/usr/bin/env bash
# Service+Endpoints: jenkins:8080 и jenkins:50000 для JNLP из k3s pod.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

JENKINS_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q jenkins)" 2>/dev/null || true)"
if [[ -z "${JENKINS_IP}" ]]; then
  echo "failed to resolve jenkins container IP (is Jenkins running?)" >&2
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

echo "jenkins Endpoints ready in k3s (default namespace)"
