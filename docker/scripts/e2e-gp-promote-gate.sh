#!/usr/bin/env bash
# GP promote blocked by draft pin → publish component → promote succeeds.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_API_KEY:-dev-local-admin-key}"
ACTOR="${COIN_PUBLISH_ACTOR:-e2e-gp-promote-gate}"
GP="${COIN_GP_E2E_NAME:-go-app}"
GP_STABLE="${COIN_GP_STABLE_VERSION:-1.0.0}"
GP_DRAFT="${COIN_GP_PROMOTE_GATE_VERSION:-1.0.3-promote-gate}"
BM_TYPE="branching-model"
BM_NAME="trunk-based"
BM_DRAFT="${COIN_BM_GATE_VERSION:-9.9.9-bm-gate}"
AUTH=(-H "X-API-Key: ${KEY}" -H "Content-Type: application/json")
MODEL_YAML="${REPO_ROOT}/coin-branching-models/models/${BM_NAME}/model.yaml"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
for cmd in curl jq python3; do need "${cmd}"; done

api_post_code() {
  local path="$1" body="$2"
  curl -sS -o /tmp/coin-e2e-body.json -w '%{http_code}' -X POST "${API}${path}" "${AUTH[@]}" -d "${body}"
}

echo "==> coin-api ready"
curl -fsS "${API}/ready" >/dev/null

comp="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_STABLE}" "${AUTH[@]}" | jq -c '.composition')"
agent="$(echo "${comp}" | jq -r '.[] | select(.type=="agent") | .version')"
gc="$(echo "${comp}" | jq -r '.[] | select(.type=="gp-content") | .version')"

echo "==> branching-model draft ${BM_NAME}@${BM_DRAFT}"
curl -fsS -X POST "${API}/v1/admin/components" "${AUTH[@]}" \
  -d "$(jq -n --arg n "${BM_NAME}" --arg a "${ACTOR}" '{type: "branching-model", name: $n, actor: $a}')" >/dev/null 2>&1 || true
curl -fsS -X POST "${API}/v1/admin/components/${BM_TYPE}/${BM_NAME}/versions/drafts" "${AUTH[@]}" \
  -d "$(jq -n --arg v "${BM_DRAFT}" --arg a "${ACTOR}" '{version: $v, actor: $a}')" >/dev/null 2>&1 || true

jq -n --rawfile b "${MODEL_YAML}" '{body: $b}' | curl -fsS -X PUT \
  "${API}/v1/admin/components/${BM_TYPE}/${BM_NAME}/versions/${BM_DRAFT}/artifacts/model.yaml" \
  "${AUTH[@]}" -d @- >/dev/null

MANIFEST_SUBSET="$(python3 - "${MODEL_YAML}" <<'PY'
import json, pathlib, sys, yaml
doc = yaml.safe_load(pathlib.Path(sys.argv[1]).read_text())
print(json.dumps({
    "branching": {
        "name": doc["name"],
        "branches": doc["branches"],
    }
}))
PY
)"

draft_body="$(jq -n --arg ver "${GP_DRAFT}" --arg agent "${agent}" --arg gc "${gc}" --arg bm "${BM_DRAFT}" --arg a "${ACTOR}" \
  '{version: $ver, agentStackName: "coin-agent", gpContentName: "go-app", branchingModelName: "trunk-based", composition: {agent: $agent, "gp-content": $gc, "branching-model": $bm}, actor: $a}')"

echo "==> GP draft with draft branching-model pin"
code="$(api_post_code "/v1/admin/golden-paths/${GP}/drafts" "${draft_body}")"
if [[ "${code}" != "201" && "${code}" != "409" ]]; then
  echo "FAIL: create GP draft HTTP ${code}: $(cat /tmp/coin-e2e-body.json)" >&2
  exit 1
fi

echo "==> promote GP (expect 409 blockingPins)"
code="$(api_post_code "/v1/admin/golden-paths/${GP}/versions/${GP_DRAFT}/promote?actor=${ACTOR}" '{}')"
body="$(cat /tmp/coin-e2e-body.json)"
if [[ "${code}" != "409" ]]; then
  echo "FAIL: expected HTTP 409, got ${code}: ${body}" >&2
  exit 1
fi
if ! echo "${body}" | jq -e '.blockingPins | length > 0' >/dev/null; then
  echo "FAIL: expected blockingPins: ${body}" >&2
  exit 1
fi
echo "OK: promote blocked"

echo "==> register + promote branching-model"
code="$(api_post_code "/v1/admin/components/${BM_TYPE}/${BM_NAME}/versions/${BM_DRAFT}/register-package" \
  "$(jq -n --argjson manifest "${MANIFEST_SUBSET}" --arg a "${ACTOR}" '{manifest: $manifest, actor: $a}')")"
if [[ "${code}" != "200" ]]; then
  echo "FAIL: register HTTP ${code}: $(cat /tmp/coin-e2e-body.json)" >&2
  exit 1
fi
code="$(api_post_code "/v1/admin/components/${BM_TYPE}/${BM_NAME}/versions/${BM_DRAFT}/promote" \
  "$(jq -n --arg a "${ACTOR}" '{actor: $a}')")"
if [[ "${code}" != "200" ]]; then
  echo "FAIL: promote component HTTP ${code}: $(cat /tmp/coin-e2e-body.json)" >&2
  exit 1
fi

echo "==> promote GP (expect 200)"
code="$(api_post_code "/v1/admin/golden-paths/${GP}/versions/${GP_DRAFT}/promote?actor=${ACTOR}" '{}')"
if [[ "${code}" != "200" ]]; then
  echo "FAIL: promote GP HTTP ${code}: $(cat /tmp/coin-e2e-body.json)" >&2
  exit 1
fi

published_ver="${GP_DRAFT%-draft}"
gp_status="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${published_ver}" "${AUTH[@]}" 2>/dev/null | jq -r '.status' || echo "")"
if [[ "${gp_status}" != "published" ]]; then
  gp_status="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_DRAFT}" "${AUTH[@]}" | jq -r '.status')"
fi
if [[ "${gp_status}" != "published" ]]; then
  echo "FAIL: GP not published (${gp_status})" >&2
  exit 1
fi

echo "OK: GP promote gate E2E (${GP})"
