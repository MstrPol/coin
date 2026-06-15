#!/usr/bin/env bash
# E2E checks for jenkins-lib-nexus model (JL-08).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
TOKEN="${COIN_API_TOKEN:-dev-local-token}"
GP="${COIN_E2E_GP:-go-app}"
VER="${COIN_E2E_VERSION:-1.0.1}"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
need curl
need jq

echo "==> JL-08 E2E: ${GP}@${VER}"

curl -fsS "${API}/ready" >/dev/null

echo "==> resolve manifest (no orchestration.bundle)"
manifest="$(curl -fsS "${API}/v1/golden-paths/${GP}/versions/${VER}/manifest" \
  -H "Authorization: Bearer ${TOKEN}")"
echo "${manifest}" | jq -e '.manifestHash | startswith("sha256:")'
echo "${manifest}" | jq -e '.pipeline.stages | length > 0'
if echo "${manifest}" | jq -e '.orchestration' >/dev/null 2>&1; then
  echo "FAIL: manifest must not contain orchestration" >&2
  exit 1
fi
echo "${manifest}" | jq -e '.capabilities.deliverables | index("image")'

echo "==> GP composition uses coin-lib@1.0.0"
curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${VER}" \
  -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
  | jq -e '.composition[] | select(.type=="lib" and .name=="coin-lib" and .version=="1.0.0")'

echo "==> registry: no pipeline-bundle"
if curl -fsS "${API}/v1/admin/components" -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
  | jq -e '.items[] | select(.type=="pipeline-bundle")' >/dev/null 2>&1; then
  echo "FAIL: pipeline-bundle still in registry" >&2
  exit 1
fi

echo "==> lib + gp-content registered"
curl -fsS "${API}/v1/admin/components/lib/coin-lib/versions" \
  -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
  | jq -e '.items | length > 0'
curl -fsS "${API}/v1/admin/components/gp-content/go-app/versions" \
  -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
  | jq -e '.items | length > 0'

echo "OK: jenkins-lib E2E API checks passed"
