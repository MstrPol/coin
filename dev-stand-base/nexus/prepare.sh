#!/usr/bin/env bash
# Инициализация Nexus dev-stand: admin, repos, пользователь coin.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
set -a && source "${ROOT}/.env" && set +a

NEXUS_HTTP_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_CACHE_REPO="${NEXUS_DOCKER_CACHE_REPO:-coin-cache}"
NEXUS_DOCKER_GROUP="${NEXUS_DOCKER_GROUP:-coin-docker-group}"
NEXUS_MAVEN_RELEASES="${NEXUS_MAVEN_RELEASES:-maven-releases}"
NEXUS_MAVEN_SNAPSHOTS="${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}"
NEXUS_ADMIN_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"
auth="admin:${NEXUS_ADMIN_PASSWORD}"

nexus_api() {
  curl -sf -u "${auth}" "$@"
}

# Hosted Docker: optional httpPort (connector). Community Edition: push только в hosted;
# group + writableMember — Pro-only, поэтому :8082 вешаем на coin-docker, не на group.
docker_hosted_payload() {
  local name=$1
  local http_port=${2:-}
  local docker_block
  if [[ -n "${http_port}" ]]; then
    docker_block="$(cat <<EOF
  "docker": {
    "v1Enabled": false,
    "forceBasicAuth": false,
    "httpPort": ${http_port}
  }
EOF
)"
  else
    docker_block="$(cat <<EOF
  "docker": {
    "v1Enabled": false,
    "forceBasicAuth": false
  }
EOF
)"
  fi
  cat <<EOF
{
  "name": "${name}",
  "online": true,
  "storage": {
    "blobStoreName": "default",
    "strictContentTypeValidation": true,
    "writePolicy": "ALLOW"
  },
${docker_block}
}
EOF
}

ensure_docker_hosted() {
  local name=$1
  local http_port=${2:-}
  local payload
  payload="$(docker_hosted_payload "${name}" "${http_port}")"
  if nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/${name}" >/dev/null 2>&1; then
    if [[ -n "${http_port}" ]]; then
      echo "==> updating Docker hosted ${name} (connector :${http_port})"
    else
      echo "==> updating Docker hosted ${name} (no connector)"
    fi
    nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/docker/hosted/${name}" \
      -H "Content-Type: application/json" \
      -d "${payload}" >/dev/null
    return
  fi
  if [[ -n "${http_port}" ]]; then
    echo "==> creating Docker hosted ${name} (connector :${http_port})"
  else
    echo "==> creating Docker hosted ${name} (no connector)"
  fi
  nexus_api -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/docker/hosted" \
    -H "Content-Type: application/json" \
    -d "${payload}"
}

# Group без connector — только для UI/агрегации; push идёт в hosted :8082.
ensure_docker_group() {
  local payload
  payload="$(cat <<EOF
{
  "name": "${NEXUS_DOCKER_GROUP}",
  "online": true,
  "storage": {
    "blobStoreName": "default",
    "strictContentTypeValidation": true
  },
  "docker": {
    "v1Enabled": false,
    "forceBasicAuth": false
  },
  "group": {
    "memberNames": ["${NEXUS_DOCKER_REPO}", "${NEXUS_DOCKER_CACHE_REPO}"]
  }
}
EOF
)"
  if nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/${NEXUS_DOCKER_GROUP}" >/dev/null 2>&1; then
    echo "==> updating Docker group ${NEXUS_DOCKER_GROUP} (no connector; push → ${NEXUS_DOCKER_REPO})"
    nexus_api -X PUT "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/docker/group/${NEXUS_DOCKER_GROUP}" \
      -H "Content-Type: application/json" \
      -d "${payload}" >/dev/null
    return
  fi
  echo "==> creating Docker group ${NEXUS_DOCKER_GROUP} (no connector)"
  nexus_api -X POST "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/docker/group" \
    -H "Content-Type: application/json" \
    -d "${payload}"
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

# Порядок важен:
# 1) если group уже есть — снять с него connector (освободить :8082);
# 2) создать/обновить hosted members (иначе fresh group create → 400);
# 3) создать/обновить group с members.
if nexus_api "http://localhost:${NEXUS_HTTP_PORT}/service/rest/v1/repositories/${NEXUS_DOCKER_GROUP}" >/dev/null 2>&1; then
  ensure_docker_group
fi
ensure_docker_hosted "${NEXUS_DOCKER_CACHE_REPO}"
ensure_docker_hosted "${NEXUS_DOCKER_REPO}" "${NEXUS_DOCKER_PORT}"
ensure_docker_group

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

echo "Nexus UI:     http://localhost:${NEXUS_HTTP_PORT} (admin / ${NEXUS_ADMIN_PASSWORD})"
echo "Nexus Docker: localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO} (${NEXUS_DOCKER_USER} / ${NEXUS_DOCKER_PASSWORD})"
echo "Nexus maven2:   http://localhost:${NEXUS_HTTP_PORT}/repository/${NEXUS_MAVEN_RELEASES}/coin/..."
echo "Nexus pointers: http://localhost:${NEXUS_HTTP_PORT}/repository/${NEXUS_MAVEN_SNAPSHOTS}/coin/manifest/{gp}/metadata/"
