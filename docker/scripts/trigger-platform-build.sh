#!/usr/bin/env bash
# Запуск platform pipeline: coin-cli → agent images (после bootstrap).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

JENKINS_PORT="${JENKINS_HTTP_PORT:-8080}"
JENKINS_USER="${JENKINS_ADMIN_USER:-admin}"
JENKINS_PASS="${JENKINS_ADMIN_PASSWORD:-admin}"

trigger_job() {
  local job="$1"
  local params="${2:-}"
  echo "==> triggering ${job}${params:+ ?${params}}"
  local cookie
  cookie="$(mktemp)"
  CRUMB="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" -c "${cookie}" \
    "http://localhost:${JENKINS_PORT}/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)")"
  local url="http://localhost:${JENKINS_PORT}/job/${job}/build"
  if [[ -n "${params}" ]]; then
    url="http://localhost:${JENKINS_PORT}/job/${job}/buildWithParameters?${params}"
  fi
  curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" -b "${cookie}" -c "${cookie}" \
    -H "${CRUMB}" -X POST "${url}"
  rm -f "${cookie}"
}

wait_job() {
  local job="$1"
  echo "==> waiting for ${job}"
  for _ in $(seq 1 120); do
    result="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" \
      "http://localhost:${JENKINS_PORT}/job/${job}/lastBuild/api/json?tree=building,result" \
      | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("result") or ("RUNNING" if d.get("building") else "UNKNOWN"))' 2>/dev/null || echo RUNNING)"
    if [[ "${result}" != "RUNNING" && "${result}" != "UNKNOWN" && -n "${result}" ]]; then
      echo "${job}: ${result}"
      [[ "${result}" == "SUCCESS" ]] || return 1
      return 0
    fi
    sleep 5
  done
  echo "timeout waiting for ${job}" >&2
  return 1
}

trigger_job coin-cli
wait_job coin-cli

trigger_job coin-agents "BUILD_ALL=true"
wait_job coin-agents || true

echo "platform build triggered — проверьте Jenkins UI"
