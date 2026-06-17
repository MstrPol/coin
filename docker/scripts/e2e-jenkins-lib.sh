#!/usr/bin/env bash
# E2E API checks: build-engine manifest model (BE-09).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
TOKEN="${COIN_API_TOKEN:-dev-local-token}"
GP="${COIN_E2E_GP:-go-app}"
VER="${COIN_E2E_VERSION:-1.0.0}"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
need curl
need jq

echo "==> BE-09 E2E API: ${GP}@${VER}"

curl -fsS "${API}/ready" >/dev/null

echo "==> resolve manifest (build-engine contract)"
manifest="$(curl -fsS "${API}/v1/golden-paths/${GP}/versions/${VER}/manifest" \
  -H "Authorization: Bearer ${TOKEN}")"
echo "${manifest}" | jq -e '.manifestHash | startswith("sha256:")'
echo "${manifest}" | jq -e '.build.engine == "buildkit"'
echo "${manifest}" | jq -e '.build.buildkit.dockerfile == ".coin/Containerfile"'
echo "${manifest}" | jq -e '.pipeline.stages[0].id == "validate"'
if echo "${manifest}" | jq -e '.pipeline.stages[0].script' >/dev/null 2>&1; then
  echo "FAIL: typed stages must not contain script" >&2
  exit 1
fi
if echo "${manifest}" | jq -e '.jnlp' >/dev/null 2>&1; then
  echo "FAIL: manifest must not contain jnlp" >&2
  exit 1
fi
if echo "${manifest}" | jq -e '.orchestration' >/dev/null 2>&1; then
  echo "FAIL: manifest must not contain orchestration" >&2
  exit 1
fi
echo "${manifest}" | jq -e '.runtime.image | length > 0'
echo "${manifest}" | jq -e '.capabilities.deliverables | index("image")'

echo "==> GP composition (4-slot)"
curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${VER}" \
  -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
  | jq -e '.composition[] | select(.type=="agent" and .name=="coin-agent")'
curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${VER}" \
  -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
  | jq -e '.composition[] | select(.type=="lib" and .name=="coin-lib" and .version=="1.0.0")'

echo "==> registry components"
for comp in "agent/coin-agent" "executor/coin-executor" "lib/coin-lib" "gp-content/go-app"; do
  typ="${comp%%/*}"
  name="${comp#*/}"
  curl -fsS "${API}/v1/admin/components/${typ}/${name}/versions" \
    -H "X-API-Key: ${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}" \
    | jq -e '.items | length > 0'
done

echo "OK: build-engine E2E API checks passed"
