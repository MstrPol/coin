#!/usr/bin/env bash
# E2E: trigger Jenkins demo-go-app/main and wait for SUCCESS.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

JENKINS_USER="${JENKINS_ADMIN_USER:-admin}"
JENKINS_PASS="${JENKINS_ADMIN_PASSWORD:-admin}"
JENKINS_PORT="${JENKINS_HTTP_PORT:-8080}"
JOB="${COIN_E2E_JOB:-demo-go-app}"
BRANCH="${COIN_E2E_BRANCH:-main}"
PUBLISH="${COIN_E2E_PUBLISH:-false}"
TIMEOUT_SEC="${COIN_E2E_TIMEOUT_SEC:-900}"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
need curl
need jq

jenkins_wait

if [[ "${COIN_E2E_SKIP_PRUNE:-0}" != "1" ]]; then
  COIN_PRUNE_DOCKER=1 bash "${ROOT}/scripts/prune-k3s-disk.sh" --all
fi

job_url="http://localhost:${JENKINS_PORT}/job/${JOB}/job/${BRANCH}"
api_base="${job_url}/api/json"
build_api="${job_url}/lastBuild/api/json"

before_num="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" "${build_api}" 2>/dev/null \
  | jq -r '.number // 0' || echo 0)"

echo "==> trigger ${JOB}/${BRANCH} publish=${PUBLISH} (last build #${before_num})"

cookie="$(mktemp)"
crumb="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" -c "${cookie}" \
  "http://localhost:${JENKINS_PORT}/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)")"

build_url="${job_url}/buildWithParameters?publish=${PUBLISH}"
code="$(curl -sS -o /dev/null -w '%{http_code}' -u "${JENKINS_USER}:${JENKINS_PASS}" \
  -b "${cookie}" -H "${crumb}" -X POST "${build_url}")"
rm -f "${cookie}"

if [[ "${code}" != "201" && "${code}" != "200" && "${code}" != "302" ]]; then
  echo "FAIL: Jenkins build trigger HTTP ${code} (${build_url})" >&2
  exit 1
fi

echo "==> waiting for build to finish (timeout ${TIMEOUT_SEC}s)"
deadline=$((SECONDS + TIMEOUT_SEC))
last_num=""
while (( SECONDS < deadline )); do
  info="$(curl -sf -u "${JENKINS_USER}:${JENKINS_PASS}" "${build_api}" 2>/dev/null || echo '{}')"
  building="$(echo "${info}" | jq -r '.building // false')"
  result="$(echo "${info}" | jq -r '.result // "null"')"
  num="$(echo "${info}" | jq -r '.number // empty')"
  url="$(echo "${info}" | jq -r '.url // empty')"
  if [[ -n "${num}" && "${num}" != "${last_num}" ]]; then
    echo "    build #${num} started: ${url}"
    last_num="${num}"
  fi
  if [[ -n "${num}" && "${num}" -gt "${before_num}" && "${building}" == "false" && "${result}" != "null" ]]; then
    echo "==> build #${num} result=${result}"
    if [[ "${result}" == "SUCCESS" ]]; then
      echo "OK: ${JOB}/${BRANCH} SUCCESS"
      exit 0
    fi
    echo "FAIL: ${JOB}/${BRANCH} ${result} — ${url}" >&2
    exit 1
  fi
  sleep 5
done

echo "FAIL: timeout waiting for ${JOB}/${BRANCH}" >&2
exit 1
