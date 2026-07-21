#!/usr/bin/env bash
# BML-4.3: canary GP resolve returns manifest.branching from draft pin (PG, no Nexus package).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_API_KEY:-dev-local-admin-key}"
TOKEN="${COIN_API_TOKEN:-dev-local-token}"
ACTOR="${COIN_PUBLISH_ACTOR:-e2e-branching-canary-resolve}"
COMP_TYPE="branching-model"
MODEL="trunk-based"
BM_VER="1.1.0-canary"
GP="go-app"
GP_STABLE="1.0.0"
GP_CANARY="1.0.1"
AUTH=(-H "X-API-Key: ${KEY}" -H "Content-Type: application/json")

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
for cmd in curl jq python3; do need "${cmd}"; done

api_post() {
  local path="$1" body="$2"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X POST "${API}${path}" "${AUTH[@]}" -d "${body}")"
  if [[ "${code}" != "201" && "${code}" != "200" && "${code}" != "409" ]]; then
    echo "POST ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  cat "${tmp}"
  rm -f "${tmp}"
}

echo "==> coin-api ready"
curl -fsS "${API}/ready" >/dev/null

echo "==> ensure GP ${GP}@${GP_STABLE} exists"
curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_STABLE}" "${AUTH[@]}" >/dev/null

echo "==> branching-model ${MODEL}@${BM_VER} draft (PG-only manifest)"
api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/drafts" \
  "$(jq -n --arg v "${BM_VER}" --arg a "${ACTOR}" '{version: $v, actor: $a}')" >/dev/null 2>&1 || true

status="$(curl -fsS "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${BM_VER}" "${AUTH[@]}" | jq -r '.status' 2>/dev/null || echo "")"
if [[ "${status}" != "draft" && "${status}" != "published" ]]; then
  MODEL_YAML="$(mktemp)"
  python3 - "${ROOT}/testdata/branching-models/${MODEL}/model.yaml" "${MODEL_YAML}" <<'PY'
import pathlib, sys, yaml
src, dst = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
doc = yaml.safe_load(src.read_text())
for br in doc.get("branches", []):
    if br.get("name") == "feature":
        br["publish"] = True
dst.write_text(yaml.dump(doc, sort_keys=False))
PY

  jq -n --rawfile b "${MODEL_YAML}" '{body: $b}' | curl -fsS -X PUT \
    "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${BM_VER}/artifacts/model.yaml" \
    "${AUTH[@]}" -d @- >/dev/null
  rm -f "${MODEL_YAML}"

  MANIFEST_SUBSET="$(python3 - "${ROOT}/testdata/branching-models/${MODEL}/model.yaml" <<'PY'
import json, pathlib, sys, yaml
doc = yaml.safe_load(pathlib.Path(sys.argv[1]).read_text())
for br in doc.get("branches", []):
    if br.get("name") == "feature":
        br["publish"] = True
print(json.dumps({
    "branching": {
        "name": doc["name"],
        "branches": doc["branches"],
    }
}))
PY
)"

  api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${BM_VER}/register-package" \
    "$(jq -n --argjson manifest "${MANIFEST_SUBSET}" --arg a "${ACTOR}" '{manifest: $manifest, actor: $a}')" >/dev/null

  if curl -fsS "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${BM_VER}" "${AUTH[@]}" \
    | jq -e '.contentRef.package.url // empty | length > 0' >/dev/null; then
    echo "FAIL: draft branching model must not have package.url" >&2
    exit 1
  fi
fi

echo "==> publish GP ${GP}@${GP_CANARY} with canary branching-model@${BM_VER}"
comp="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_STABLE}" "${AUTH[@]}" | jq -c '.composition')"
destinations="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_STABLE}" "${AUTH[@]}" | jq -c '.destinations')"
agent="$(echo "${comp}" | jq -r '.[] | select(.type=="agent") | .version')"
content="$(echo "${comp}" | jq -r '.[] | select(.type=="gp-content") | .version')"
api_post "/v1/admin/golden-paths/${GP}/versions" "$(jq -n \
  --arg ver "${GP_CANARY}" \
  --argjson destinations "${destinations}" \
  --arg agent "${agent}" --arg content "${content}" --arg bm "${BM_VER}" --arg a "${ACTOR}" \
  '{version: $ver, destinations: $destinations, agentStackName: "coin-agent", gpContentName: "go-app", branchingModelName: "trunk-based", composition: {agent: $agent, "gp-content": $content, "branching-model": $bm}, actor: $a}')" >/dev/null

curl -fsS -X PATCH "${API}/v1/admin/golden-paths/${GP}/catalog" "${AUTH[@]}" \
  -d "$(jq -n --arg stable "${GP_STABLE}" --arg canary "${GP_CANARY}" --arg a "${ACTOR}" \
    '{latest: $stable, latestCanary: $canary, minimum: $stable, actor: $a}')" >/dev/null

echo "==> resolve canary pin (forceChannel=canary)"
preview="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/resolve-preview?pin=*&forceChannel=canary" \
  -H "X-API-Key: ${KEY}")"
channel="$(echo "${preview}" | jq -r '.channel')"
feature_publish="$(echo "${preview}" | jq -r '.manifest.branching.branches[] | select(.name=="feature") | .publish')"
bm_name="$(echo "${preview}" | jq -r '.manifest.branching.name')"
bm_ver="$(echo "${preview}" | jq -r '.manifest.branching.version // empty')"

if [[ "${channel}" != "canary" ]]; then
  echo "FAIL: expected channel canary, got ${channel}" >&2
  exit 1
fi
if [[ "${feature_publish}" != "true" ]]; then
  echo "FAIL: expected canary feature.publish=true from PG ref, got ${feature_publish}" >&2
  echo "${preview}" | jq '.manifest.branching' >&2
  exit 1
fi
if [[ "${bm_name}" != "${MODEL}" ]]; then
  echo "FAIL: unexpected branching.name ${bm_name}" >&2
  exit 1
fi

echo "==> stable resolve unchanged (feature.publish=false)"
stable_feature="$(curl -fsS "${API}/v1/golden-paths/${GP}/versions/${GP_STABLE}/manifest" \
  -H "Authorization: Bearer ${TOKEN}" | jq -r '.branching.branches[] | select(.name=="feature") | .publish')"
if [[ "${stable_feature}" != "false" ]]; then
  echo "FAIL: stable branching feature.publish should remain false" >&2
  exit 1
fi

echo "OK: canary resolve manifest.branching from PG (${bm_name}@${BM_VER:-canary}, feature.publish=${feature_publish})"
