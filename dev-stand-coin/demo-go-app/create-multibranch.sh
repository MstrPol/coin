#!/usr/bin/env bash
# Jenkins multibranch job demo-go-app → Gitea coin/demo-go-app.
set -euo pipefail

stand_coin="$(cd "$(dirname "$0")/.." && pwd)"
stand_base="$(cd "${stand_coin}/../dev-stand-base" && pwd)"
job_dir="$(cd "$(dirname "$0")" && pwd)"
repo_name="${DEMO_GO_APP_REPO:-demo-go-app}"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

JENKINS_HTTP_PORT="${JENKINS_HTTP_PORT:-8080}"
JENKINS_ADMIN_USER="${JENKINS_ADMIN_USER:-admin}"
JENKINS_ADMIN_PASSWORD="${JENKINS_ADMIN_PASSWORD:-admin}"

echo "==> waiting for Jenkins"
for _ in $(seq 1 60); do
  curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" \
    "http://localhost:${JENKINS_HTTP_PORT}/login" && break
  sleep 3
done
curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" \
  "http://localhost:${JENKINS_HTTP_PORT}/login" || {
  echo "Jenkins не отвечает на http://localhost:${JENKINS_HTTP_PORT}" >&2
  exit 1
}

jenkins_container="$(
  docker ps --filter "publish=${JENKINS_HTTP_PORT}" --format '{{.Names}}' | head -1
)"
[[ -n "${jenkins_container}" ]] || {
  echo "Jenkins-контейнер не найден (порт ${JENKINS_HTTP_PORT})" >&2
  exit 1
}

echo "==> Jenkins multibranch ${repo_name} (CASC)"
docker exec -i -u 0 "${jenkins_container}" sh -c \
  "cat > /var/jenkins_home/casc-config/48-demo-go-app.yaml && chown jenkins:jenkins /var/jenkins_home/casc-config/48-demo-go-app.yaml" \
  < "${job_dir}/casc-multibranch.yaml"

cookie="$(mktemp)"
crumb="$(curl -sf -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" -c "${cookie}" \
  "http://localhost:${JENKINS_HTTP_PORT}/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)")"

http_code="$(curl -sf -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" -b "${cookie}" -c "${cookie}" \
  -H "${crumb}" -o /dev/null -w '%{http_code}' \
  -X POST "http://localhost:${JENKINS_HTTP_PORT}/configuration-as-code/reload")"

[[ "${http_code}" == "200" || "${http_code}" == "302" ]] || {
  rm -f "${cookie}"
  echo "Jenkins CASC reload failed (HTTP ${http_code})" >&2
  exit 1
}

# Branch Indexing
echo "==> scan branches"
crumb="$(curl -sf -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" -c "${cookie}" \
  "http://localhost:${JENKINS_HTTP_PORT}/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)")"
curl -sf -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" -b "${cookie}" \
  -H "${crumb}" \
  -X POST "http://localhost:${JENKINS_HTTP_PORT}/job/${repo_name}/build" \
  >/dev/null || true
rm -f "${cookie}"

echo "OK: http://localhost:${JENKINS_HTTP_PORT}/job/${repo_name}/"
echo "    перед билдом: make push-demo-go-app && make coin-lib && make publish-executor"
