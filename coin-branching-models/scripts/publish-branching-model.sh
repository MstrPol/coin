#!/usr/bin/env bash
# Bootstrap publish: branching-model draft → register-package → canary → published.
set -euo pipefail

MODEL="${1:?model name (e.g. trunk-based)}"
VERSION="${2:?version (e.g. 1.0.0)}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODEL_DIR="${ROOT}/models/${MODEL}"
MODEL_YAML="${MODEL_DIR}/model.yaml"
OUT_DIR="${ROOT}/dist"
COMP_TYPE="branching-model"

COIN_API_URL="${COIN_API_URL:-http://localhost:8090}"
API_KEY="${COIN_API_KEY:-dev-local-admin-key}"
ACTOR="${COIN_PUBLISH_ACTOR:-coin-branching-models-ci}"
AUTH=(-H "X-API-Key: ${API_KEY}" -H "Content-Type: application/json")

for cmd in curl jq python3; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

if [[ ! -f "${MODEL_YAML}" ]]; then
  echo "model not found: ${MODEL_YAML}" >&2
  exit 1
fi

mkdir -p "${OUT_DIR}"

api_post() {
  local path="$1" body="$2"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X POST "${COIN_API_URL}${path}" "${AUTH[@]}" -d "${body}")"
  if [[ "${code}" != "201" && "${code}" != "200" && "${code}" != "409" ]]; then
    echo "POST ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  rm -f "${tmp}"
}

published_version() {
  curl -fsS "${COIN_API_URL}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions" "${AUTH[@]}" 2>/dev/null \
    | jq -r --arg v "${VERSION}" '
        [.items[] | select(.status == "published" and .version == $v) | .version]
        | first // empty
      ' || true
}

if [[ -n "$(published_version)" ]]; then
  echo "==> already published ${COMP_TYPE}/${MODEL}@${VERSION}"
  exit 0
fi

echo "==> ensure component ${COMP_TYPE}/${MODEL}"
api_post "/v1/admin/components" "$(jq -n --arg t "${COMP_TYPE}" --arg n "${MODEL}" --arg a "${ACTOR}" \
  '{type: $t, name: $n, actor: $a}')"

echo "==> create draft ${VERSION}"
api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/drafts" \
  "$(jq -n --arg v "${VERSION}" --arg a "${ACTOR}" '{version: $v, actor: $a}')"

echo "==> upload artifact model.yaml"
jq -n --rawfile b "${MODEL_YAML}" '{body: $b}' | curl -fsS -X PUT \
  "${COIN_API_URL}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/artifacts/model.yaml" \
  "${AUTH[@]}" \
  -d @-

python3 - "${MODEL_YAML}" "${OUT_DIR}/manifest-subset.json" <<'PY'
import json, pathlib, sys, yaml

model_path = pathlib.Path(sys.argv[1])
out_path = pathlib.Path(sys.argv[2])
doc = yaml.safe_load(model_path.read_text())
subset = {
    "branching": {
        "name": doc["name"],
        "trunk": doc["trunk"],
        "branchTypes": doc["branchTypes"],
        "versioning": doc["versioning"],
        "publish": doc["publish"],
    }
}
out_path.write_text(json.dumps(subset))
PY

echo "==> register-package (Nexus + content_ref v2)"
register_body="$(jq -n \
  --slurpfile manifest "${OUT_DIR}/manifest-subset.json" \
  --arg a "${ACTOR}" \
  '{manifest: $manifest[0], actor: $a}')"
register_tmp="$(mktemp)"
register_code="$(curl -sS -o "${register_tmp}" -w '%{http_code}' -X POST \
  "${COIN_API_URL}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/register-package" \
  "${AUTH[@]}" -d "${register_body}")"
if [[ "${register_code}" != "200" ]]; then
  echo "register-package failed HTTP ${register_code}: $(cat "${register_tmp}")" >&2
  rm -f "${register_tmp}"
  exit 1
fi
rm -f "${register_tmp}"

echo "==> publish-canary"
api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/publish-canary" \
  "$(jq -n --arg a "${ACTOR}" '{actor: $a}')"

echo "==> promote to published"
promote_tmp="$(mktemp)"
promote_code="$(curl -sS -o "${promote_tmp}" -w '%{http_code}' -X POST \
  "${COIN_API_URL}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/promote" \
  "${AUTH[@]}" -d "$(jq -n --arg a "${ACTOR}" '{actor: $a}')")"
if [[ "${promote_code}" != "200" && "${promote_code}" != "409" ]]; then
  echo "promote failed HTTP ${promote_code}: $(cat "${promote_tmp}")" >&2
  rm -f "${promote_tmp}"
  exit 1
fi
rm -f "${promote_tmp}"

echo "==> done ${COMP_TYPE}/${MODEL}@${VERSION}"
