#!/usr/bin/env bash
# MVP-1 E2E checks: resolve manifest, Nexus blob+pointer fallback, content artifacts (PF-11).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
NEXUS_PORT="${NEXUS_HTTP_PORT:-8081}"
NEXUS="${NEXUS_URL:-http://localhost:${NEXUS_PORT}}"
REPO="${NEXUS_MANIFEST_REPO:-coin-manifests}"
BASE="${NEXUS}/repository/${REPO}"
GP="${COIN_E2E_GP:-go-app}"
VER="${COIN_E2E_VERSION:-1.0.0}"
PIN_PATH="%3D${VER}"

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1" >&2; exit 1; }
}
need curl
need jq

# Manifest/Nexus URLs use docker-internal hostnames; rewrite for host-side curl.
host_url() {
  local url="$1"
  url="${url//http:\/\/nexus:8081/http://localhost:${NEXUS_PORT}}"
  url="${url//http:\/\/nexus:8082/http://localhost:${NEXUS_DOCKER_PORT:-8082}}"
  echo "${url}"
}

echo "==> PF-11 E2E: ${GP}@${VER}"

echo "==> coin-api /ready"
curl -fsS "${API}/ready" >/dev/null

echo "==> resolve manifest (warms Nexus blob + pointer + content)"
manifest="$(curl -fsS "${API}/v1/golden-paths/${GP}/versions/${VER}/manifest")"
echo "${manifest}" | jq -e '.manifestHash | startswith("sha256:")' >/dev/null
echo "${manifest}" | jq -e '.orchestration.url | length > 0' >/dev/null
echo "${manifest}" | jq -e '.pipeline.stages[0].script.url | length > 0' >/dev/null
if echo "${manifest}" | grep -q gitRef; then
  echo "FAIL: manifest still contains gitRef" >&2
  exit 1
fi

echo "==> Nexus pointer (exact pin =${VER})"
ptr="$(curl -fsS "${BASE}/pointers/${GP}/${PIN_PATH}.json")"
blob_url="$(echo "${ptr}" | jq -r .blobUrl)"
expected_hash="$(echo "${ptr}" | jq -r .manifestHash)"
[[ -n "${blob_url}" && "${blob_url}" != "null" ]]
blob_url="$(host_url "${blob_url}")"

echo "==> Nexus blob + manifestHash verify"
actual_hash="$(curl -fsS "${blob_url}" | jq -r .manifestHash)"
[[ "${expected_hash}" == "${actual_hash}" ]]

echo "==> Nexus content artifact (test.sh)"
curl -fsS "${BASE}/content/${GP}/${VER}/scripts/test.sh" | grep -q coin

echo "==> orchestration groovy reachable"
orch_url="$(host_url "$(echo "${manifest}" | jq -r .orchestration.url)")"
curl -fsS "${orch_url}" | grep -q coinPipeline

echo "==> API-down fallback simulation (pointer → blob only)"
sim_hash="$(curl -fsS "${blob_url}" | jq -r .manifestHash)"
[[ "${sim_hash}" == "${expected_hash}" ]]

echo "OK: MVP-1 E2E checks passed (${GP}@${VER})"
