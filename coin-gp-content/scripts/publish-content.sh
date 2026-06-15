#!/usr/bin/env bash
# Build gp-content zip, upload to Nexus, register in coin-api.
set -euo pipefail

STACK="${1:?gp-content name (e.g. go-app)}"
VERSION="${2:?version (e.g. 1.0.0)}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STACK_DIR="${ROOT}/stacks/${STACK}"
OUT_DIR="${ROOT}/dist"
PAYLOAD_DIR="${OUT_DIR}/.publish-payload"
ZIP="${OUT_DIR}/gp-content-${STACK}-${VERSION}.zip"

COIN_API_URL="${COIN_API_URL:-http://localhost:8090}"
API_KEY="${COIN_API_KEY:-dev-local-admin-key}"
# shellcheck source=lib/maven-url.sh
source "$(dirname "$0")/lib/maven-url.sh"

for cmd in zip curl jq python3; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

if [[ ! -d "${STACK_DIR}" ]]; then
  echo "stack not found: ${STACK_DIR}" >&2
  exit 1
fi

mkdir -p "${OUT_DIR}"
rm -rf "${PAYLOAD_DIR}" "${ZIP}"

(
  cd "${STACK_DIR}"
  zip -qr "${ZIP}" content.yaml scripts dockerfiles schemas
)

SHA256="sha256:$(sha256sum "${ZIP}" | awk '{print $1}')"
CONTENT_URL="$(gp_content_url "${STACK}" "${VERSION}")"

nexus_upload() {
  local url="$1" file="$2"
  local auth="${NEXUS_USER:-admin}:${NEXUS_PASSWORD:-admin123}"
  local code
  code="$(curl -s -o /dev/null -w '%{http_code}' -u "${auth}" -I "${url}" || true)"
  if [[ "${code}" == "200" ]]; then
    echo "==> already in Nexus, skip upload: ${url}"
    return 0
  fi
  echo "==> upload ${file} -> ${url}"
  local body tmp
  tmp="$(mktemp)"
  code="$(curl -sS -u "${auth}" -o "${tmp}" -w '%{http_code}' --upload-file "${file}" "${url}")"
  if [[ "${code}" == "201" || "${code}" == "200" || "${code}" == "204" ]]; then
    return 0
  fi
  body="$(cat "${tmp}")"
  rm -f "${tmp}"
  if [[ "${code}" == "400" && "${body}" == *"already exists"* ]]; then
    echo "==> already in Nexus (immutable repo), continue: ${url}"
    return 0
  fi
  echo "nexus upload failed HTTP ${code}: ${body}" >&2
  return 1
}

nexus_upload "${CONTENT_URL}" "${ZIP}"

mkdir -p "${PAYLOAD_DIR}"
python3 - "${STACK_DIR}" "${PAYLOAD_DIR}" "${CONTENT_URL}" "${SHA256}" <<'PY'
import json, hashlib, pathlib, sys, yaml

stack = pathlib.Path(sys.argv[1])
out = pathlib.Path(sys.argv[2])
content_url = sys.argv[3]
content_sha = sys.argv[4]
content = yaml.safe_load((stack / "content.yaml").read_text())

def sha(p: pathlib.Path) -> str:
    return f"sha256:{hashlib.sha256(p.read_bytes()).hexdigest()}"

stages = []
for st in content.get("stages", []):
    key = st["artifactKey"]
    item = {"name": st["name"], "artifactKey": key, "sha256": sha(stack / key)}
    if st.get("when"):
        item["when"] = st["when"]
    stages.append(item)

vs = content["validateSchema"]["artifactKey"]
df = content["dockerfileTemplate"]["artifactKey"]

cref = {
    "stages": stages,
    "validateSchema": {"artifactKey": vs, "sha256": sha(stack / vs)},
    "dockerfileTemplate": {"artifactKey": df, "sha256": sha(stack / df)},
}
controls = content.get("controls") or {}
capabilities = content.get("capabilities") or {}
meta = {
    "url": content_url,
    "sha256": content_sha,
    "buildControls": controls,
    "capabilities": capabilities,
}
keys = [s["artifactKey"] for s in stages] + [vs, df]

(out / "content-ref.json").write_text(json.dumps(cref))
(out / "metadata.json").write_text(json.dumps(meta))
(out / "keys.json").write_text(json.dumps(keys))
PY

echo "==> register gp-content/${STACK}@${VERSION}"
register_body="$(jq -n \
  --arg ver "${VERSION}" \
  --slurpfile meta "${PAYLOAD_DIR}/metadata.json" \
  --slurpfile cref "${PAYLOAD_DIR}/content-ref.json" \
  '{version: $ver, metadata: $meta[0], contentRef: $cref[0], actor: "gp-content-ci"}')"
register_tmp="$(mktemp)"
register_code="$(curl -sS -o "${register_tmp}" -w '%{http_code}' -X POST \
  "${COIN_API_URL}/v1/admin/components/gp-content/${STACK}/versions" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d "${register_body}")"
if [[ "${register_code}" != "201" && "${register_code}" != "409" ]]; then
  echo "coin-api register failed HTTP ${register_code}: $(cat "${register_tmp}")" >&2
  rm -f "${register_tmp}"
  exit 1
fi
if [[ "${register_code}" == "409" ]]; then
  echo "==> version already registered in coin-api, continue"
fi
rm -f "${register_tmp}"

jq -r '.[]' "${PAYLOAD_DIR}/keys.json" | while IFS= read -r key; do
  file="${STACK_DIR}/${key}"
  enc_key="$(jq -rn --arg k "${key}" '$k|@uri')"
  echo "==> artifact ${key}"
  jq -n --rawfile b "${file}" '{body: $b}' | curl -fsS -X PUT \
    "${COIN_API_URL}/v1/admin/components/gp-content/${STACK}/versions/${VERSION}/artifacts/${enc_key}" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: ${API_KEY}" \
    -d @-
done

rm -rf "${PAYLOAD_DIR}"
echo "==> done gp-content/${STACK}@${VERSION}"
