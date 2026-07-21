#!/usr/bin/env bash
# BML-3.3: branching-model lifecycle — PG canary, Nexus on promote.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_API_KEY:-dev-local-admin-key}"
ACTOR="${COIN_PUBLISH_ACTOR:-e2e-branching-lifecycle}"
COMP_TYPE="branching-model"
MODEL="trunk-based"
VERSION="${COIN_BML_E2E_VERSION:-9.9.9-bml-e2e}"
AUTH=(-H "X-API-Key: ${KEY}" -H "Content-Type: application/json")
MODEL_YAML="${ROOT}/testdata/branching-models/${MODEL}/model.yaml"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
for cmd in curl jq python3; do need "${cmd}"; done

if [[ ! -f "${MODEL_YAML}" ]]; then
  echo "model not found: ${MODEL_YAML}" >&2
  exit 1
fi

echo "==> coin-api ready"
curl -fsS "${API}/ready" >/dev/null

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

content_ref_field() {
  local field="$1"
  curl -fsS "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}" "${AUTH[@]}" \
    | jq -r ".contentRef.${field} // empty"
}

has_package_url() {
  [[ -n "$(content_ref_field package.url)" ]]
}

cleanup_version() {
  echo "==> cleanup skipped (immutable versions); use fresh VERSION if re-run"
}

trap cleanup_version EXIT

echo "==> ensure component + draft ${VERSION}"
api_post "/v1/admin/components" "$(jq -n --arg t "${COMP_TYPE}" --arg n "${MODEL}" --arg a "${ACTOR}" \
  '{type: $t, name: $n, actor: $a}')" >/dev/null
api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/drafts" \
  "$(jq -n --arg v "${VERSION}" --arg a "${ACTOR}" '{version: $v, actor: $a}')" >/dev/null

echo "==> upload model.yaml"
jq -n --rawfile b "${MODEL_YAML}" '{body: $b}' | curl -fsS -X PUT \
  "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/artifacts/model.yaml" \
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

echo "==> register-package (expect PG-only, no package.url)"
register_out="$(api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/register-package" \
  "$(jq -n --argjson manifest "${MANIFEST_SUBSET}" --arg a "${ACTOR}" '{manifest: $manifest, actor: $a}')")"
if jq -e '.packageUrl // empty | length > 0' <<<"${register_out}" >/dev/null 2>&1; then
  echo "FAIL: register returned packageUrl for branching-model" >&2
  exit 1
fi
if has_package_url; then
  echo "FAIL: content_ref has package.url after register (expected PG-only)" >&2
  exit 1
fi
echo "OK: register PG-only"

echo "==> verify draft after register"
status="$(curl -fsS "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}" "${AUTH[@]}" | jq -r '.status')"
if [[ "${status}" != "draft" ]]; then
  echo "FAIL: expected draft after register, got ${status}" >&2
  exit 1
fi
if has_package_url; then
  echo "FAIL: draft content_ref must not have package.url yet" >&2
  exit 1
fi
branching_name="$(jq -r '.contentRef.manifest.branching.name // empty' <<<"$(curl -fsS \
  "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}" "${AUTH[@]}")")"
if [[ "${branching_name}" != "${MODEL}" ]]; then
  echo "FAIL: manifest subset missing branching.name" >&2
  exit 1
fi
echo "OK: draft PG content_ref"

echo "==> promote (draft → published, Nexus package.url)"
api_post "/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}/promote" \
  "$(jq -n --arg a "${ACTOR}" '{actor: $a}')" >/dev/null

if ! has_package_url; then
  echo "FAIL: published content_ref must include package.url" >&2
  exit 1
fi
published_status="$(curl -fsS "${API}/v1/admin/components/${COMP_TYPE}/${MODEL}/versions/${VERSION}" "${AUTH[@]}" | jq -r '.status')"
if [[ "${published_status}" != "published" ]]; then
  echo "FAIL: expected published, got ${published_status}" >&2
  exit 1
fi

echo "OK: branching-model lifecycle E2E (${COMP_TYPE}/${MODEL}@${VERSION})"
