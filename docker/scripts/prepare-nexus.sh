#!/usr/bin/env bash
# Nexus: raw repo coin-cli (anonymous read) + Endpoints в k3s.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_ADMIN_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}"

echo "==> waiting for Nexus"
for i in $(seq 1 90); do
  if curl -sf "http://localhost:${NEXUS_HTTP_PORT}/" >/dev/null 2>&1; then
    break
  fi
  sleep 3
done

if [[ ! -f "${ROOT}/.nexus-admin-initialized" ]]; then
  initial="$(
    docker compose exec -T nexus cat /nexus-data/admin.password 2>/dev/null | tr -d '\r\n' || true
  )"
  if [[ -n "${initial}" ]]; then
    echo "==> setting Nexus admin password"
    curl -sf -u "admin:${initial}" -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/users/admin/change-password" \
      -H "Content-Type: text/plain" \
      --data-binary "${NEXUS_ADMIN_PASSWORD}" || true
    touch "${ROOT}/.nexus-admin-initialized"
  fi
fi

auth="admin:${NEXUS_ADMIN_PASSWORD}"

if ! curl -sf -u "${auth}" "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/coin-cli" >/dev/null 2>&1; then
  echo "==> creating raw repo coin-cli"
  curl -sf -u "${auth}" -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/raw/hosted" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "coin-cli",
      "online": true,
      "storage": {
        "blobStoreName": "default",
        "strictContentTypeValidation": false,
        "writePolicy": "ALLOW"
      }
    }'
fi

echo "==> enabling anonymous read for coin-cli"
curl -sf -u "${auth}" -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/roles" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "coin-cli-read",
    "name": "coin-cli-read",
    "description": "Anonymous read coin-cli raw repo (local dev)",
    "privileges": ["nx-repository-view-raw-coin-cli-browse", "nx-repository-view-raw-coin-cli-read"],
    "roles": []
  }' >/dev/null 2>&1 || true

curl -sf -u "${auth}" -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/anonymous" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "userId": "anonymous",
    "realmName": "NexusAuthorizingRealm",
    "roles": ["nx-anonymous", "coin-cli-read"]
  }' >/dev/null 2>&1 || true

NEXUS_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q nexus)" 2>/dev/null || true)"
if [[ -n "${NEXUS_IP}" ]]; then
  echo "==> registering nexus in k3s (Endpoints ${NEXUS_IP}:8081)"
  docker compose exec -T k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: nexus
  namespace: default
spec:
  ports:
    - port: 8081
      targetPort: 8081
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
      - port: 8081
EOF
fi

echo "Nexus coin-cli: http://localhost:${NEXUS_HTTP_PORT}/repository/coin-cli/ (admin / ${NEXUS_ADMIN_PASSWORD})"
