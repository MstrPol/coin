#!/usr/bin/env bash
# Nexus: admin password, Docker hosted repo, пользователь coin, Endpoints в k3s.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_MAVEN_RELEASES="${NEXUS_MAVEN_RELEASES:-maven-releases}"
NEXUS_MAVEN_SNAPSHOTS="${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}"
NEXUS_ADMIN_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"
auth="admin:${NEXUS_ADMIN_PASSWORD}"

nexus_api() {
  curl -sf -u "${auth}" "$@"
}

accept_nexus_eula() {
  local eula
  eula="$(nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/system/eula")"
  if python3 -c 'import json, sys; sys.exit(0 if json.load(sys.stdin).get("accepted") else 1)' <<<"${eula}"; then
    return 0
  fi
  echo "==> accepting Nexus Community Edition EULA"
  python3 -c '
import json, sys, urllib.request, base64

eula = json.load(sys.stdin)
payload = json.dumps({"accepted": True, "disclaimer": eula["disclaimer"]}).encode()
auth_header = base64.b64encode(sys.argv[1].encode()).decode()
req = urllib.request.Request(
    sys.argv[2],
    data=payload,
    method="POST",
    headers={
        "Authorization": f"Basic {auth_header}",
        "Content-Type": "application/json",
    },
)
with urllib.request.urlopen(req) as resp:
    if resp.status not in (200, 204):
        raise SystemExit(f"EULA accept failed: HTTP {resp.status}")
' "${auth}" "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/system/eula" <<<"${eula}"
}

echo "==> waiting for Nexus"
for _ in $(seq 1 90); do
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

accept_nexus_eula

echo "==> anonymous read: ${NEXUS_MAVEN_RELEASES} + ${NEXUS_MAVEN_SNAPSHOTS} (CI curl без creds)"
nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/anonymous" \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "userId": "anonymous", "realmName": "NexusAuthorizingRealm"}' >/dev/null

nexus_api -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/roles" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"coin-maven-reader\",
    \"name\": \"coin-maven-reader\",
    \"description\": \"Anonymous read maven-releases/snapshots (local dev CI)\",
    \"privileges\": [
      \"nx-repository-view-maven2-${NEXUS_MAVEN_RELEASES}-browse\",
      \"nx-repository-view-maven2-${NEXUS_MAVEN_RELEASES}-read\",
      \"nx-repository-view-maven2-${NEXUS_MAVEN_SNAPSHOTS}-browse\",
      \"nx-repository-view-maven2-${NEXUS_MAVEN_SNAPSHOTS}-read\"
    ],
    \"roles\": []
  }" >/dev/null 2>&1 || true

nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/users/anonymous" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "anonymous",
    "firstName": "Anonymous",
    "lastName": "User",
    "emailAddress": "anonymous@local",
    "status": "active",
    "roles": ["coin-maven-reader", "nx-anonymous"]
  }' >/dev/null 2>&1 || true

echo "==> enabling Docker Bearer Token realm"
active_realms="$(nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/realms/active")"
if ! echo "${active_realms}" | grep -q 'DockerToken'; then
  updated_realms="$(python3 -c '
import json, sys
realms = json.load(sys.stdin)
if "DockerToken" not in realms:
    realms.append("DockerToken")
print(json.dumps(realms))
' <<<"${active_realms}")"
  nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/realms/active" \
    -H "Content-Type: application/json" \
    -d "${updated_realms}" >/dev/null
fi

if ! nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/${NEXUS_DOCKER_REPO}" >/dev/null 2>&1; then
  echo "==> creating Nexus Docker hosted repo ${NEXUS_DOCKER_REPO} (connector :${NEXUS_DOCKER_PORT})"
  nexus_api -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/docker/hosted" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"${NEXUS_DOCKER_REPO}\",
      \"online\": true,
      \"storage\": {
        \"blobStoreName\": \"default\",
        \"strictContentTypeValidation\": true,
        \"writePolicy\": \"ALLOW\"
      },
      \"docker\": {
        \"v1Enabled\": false,
        \"forceBasicAuth\": false,
        \"httpPort\": ${NEXUS_DOCKER_PORT}
      }
    }"
fi

echo "==> maven-snapshots: MIXED version policy (manifest pointers .../metadata/)"
nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/maven/hosted/${NEXUS_MAVEN_SNAPSHOTS}" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"${NEXUS_MAVEN_SNAPSHOTS}\",
    \"online\": true,
    \"storage\": {
      \"blobStoreName\": \"default\",
      \"strictContentTypeValidation\": true,
      \"writePolicy\": \"ALLOW\"
    },
    \"maven\": {
      \"versionPolicy\": \"MIXED\",
      \"layoutPolicy\": \"STRICT\"
    },
    \"component\": {
      \"proprietaryComponents\": false
    }
  }" >/dev/null 2>&1 || echo "    WARN: could not update ${NEXUS_MAVEN_SNAPSHOTS} (pointers may fail)"

echo "==> configuring role ${NEXUS_DOCKER_USER} for ${NEXUS_DOCKER_REPO}"
nexus_api -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/roles" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"coin-docker-writer\",
    \"name\": \"coin-docker-writer\",
    \"description\": \"Push/pull ${NEXUS_DOCKER_REPO} (local dev)\",
    \"privileges\": [
      \"nx-repository-view-docker-${NEXUS_DOCKER_REPO}-browse\",
      \"nx-repository-view-docker-${NEXUS_DOCKER_REPO}-read\",
      \"nx-repository-view-docker-${NEXUS_DOCKER_REPO}-add\",
      \"nx-repository-view-docker-${NEXUS_DOCKER_REPO}-edit\"
    ],
    \"roles\": []
  }" >/dev/null 2>&1 || true

if nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/users?userId=${NEXUS_DOCKER_USER}" \
  | grep -q '"userId"'; then
  nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/users/${NEXUS_DOCKER_USER}/change-password" \
    -H "Content-Type: text/plain" \
    --data-binary "${NEXUS_DOCKER_PASSWORD}" >/dev/null
else
  nexus_api -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/security/users" \
    -H "Content-Type: application/json" \
    -d "{
      \"userId\": \"${NEXUS_DOCKER_USER}\",
      \"firstName\": \"Coin\",
      \"lastName\": \"Dev\",
      \"emailAddress\": \"coin@local\",
      \"password\": \"${NEXUS_DOCKER_PASSWORD}\",
      \"status\": \"active\",
      \"roles\": [\"coin-docker-writer\"]
    }" >/dev/null
fi

echo "==> waiting for Docker connector :${NEXUS_DOCKER_PORT}"
for _ in $(seq 1 60); do
  code="$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${NEXUS_DOCKER_PORT}/v2/")"
  if [[ "${code}" == "200" || "${code}" == "401" ]]; then
    break
  fi
  sleep 3
done

chmod +x "${ROOT}/scripts/sync-k3s-registries.sh" "${ROOT}/scripts/register-stack-endpoints.sh"
"${ROOT}/scripts/sync-k3s-registries.sh"

NEXUS_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$(docker compose ps -q nexus)" 2>/dev/null || true)"
if [[ -n "${NEXUS_IP}" ]]; then
  "${ROOT}/scripts/register-stack-endpoints.sh"
fi

echo "Nexus UI:     http://localhost:${NEXUS_HTTP_PORT} (admin / ${NEXUS_ADMIN_PASSWORD})"
echo "Nexus Docker: localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO} (${NEXUS_DOCKER_USER} / ${NEXUS_DOCKER_PASSWORD})"
echo "Nexus maven2:   http://localhost:${NEXUS_HTTP_PORT}/repository/${NEXUS_MAVEN_RELEASES}/coin/..."
echo "Nexus pointers: http://localhost:${NEXUS_HTTP_PORT}/repository/${NEXUS_MAVEN_SNAPSHOTS}/coin/manifest/{gp}/metadata/"
