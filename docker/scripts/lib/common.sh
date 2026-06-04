#!/usr/bin/env bash
# shellcheck disable=SC1091
set -euo pipefail

DOCKER_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

load_env() {
  cd "${DOCKER_ROOT}"
  [[ -f .env ]] && source .env
}

compose() {
  docker compose -f "${DOCKER_ROOT}/compose.yml" "$@"
}

gitea_create_repo() {
  local name="$1"
  curl -sf -u "${GITEA_USER}:${GITEA_PASSWORD}" \
    -X POST "http://localhost:${GITEA_HTTP_PORT}/api/v1/user/repos" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"${name}\",\"private\":false,\"default_branch\":\"main\"}" \
    >/dev/null 2>&1 || true
}

jenkins_wait() {
  JENKINS_PORT="${JENKINS_HTTP_PORT:-8080}"
  JENKINS_USER="${JENKINS_ADMIN_USER:-admin}"
  JENKINS_PASS="${JENKINS_ADMIN_PASSWORD:-admin}"
  for _ in $(seq 1 60); do
    if curl -sf -o /dev/null -u "${JENKINS_USER}:${JENKINS_PASS}" \
      "http://localhost:${JENKINS_PORT}/login"; then
      return 0
    fi
    sleep 3
  done
  echo "Jenkins not ready on :${JENKINS_PORT}" >&2
  exit 1
}

jenkins_casc_reload() {
  local yaml_host_path="$1"
  local yaml_container_path="$2"

  compose exec -T jenkins tee "${yaml_container_path}" < "${yaml_host_path}" >/dev/null

  local cookie
  cookie="$(mktemp)"
  local crumb
  crumb="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" -c "${cookie}" \
    "http://localhost:${JENKINS_PORT}/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)")"

  local http_code
  http_code="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" -b "${cookie}" -c "${cookie}" \
    -H "${crumb}" -o /dev/null -w '%{http_code}' \
    -X POST "http://localhost:${JENKINS_PORT}/configuration-as-code/reload")"
  rm -f "${cookie}"

  if [[ "${http_code}" != "200" && "${http_code}" != "302" ]]; then
    echo "Jenkins CASC reload failed (HTTP ${http_code})" >&2
    exit 1
  fi
}
