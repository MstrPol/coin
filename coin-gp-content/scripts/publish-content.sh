#!/usr/bin/env bash
# DEPRECATED: primary publish path — Component Studio + coin-api Admin API.
# Kept for local bootstrap / CI until GCP-4 fleet migration completes.
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
  zip_items=(content.yaml)
  [[ -d schemas ]] && zip_items+=(schemas)
  [[ -d dockerfiles ]] && zip_items+=(dockerfiles)
  [[ -f project.toml ]] && zip_items+=(project.toml)
  zip -qr "${ZIP}" "${zip_items[@]}"
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
python3 - "${STACK_DIR}" "${PAYLOAD_DIR}" "${CONTENT_URL}" "${SHA256}" "${STACK}" "${VERSION}" <<'PY'
import json, hashlib, pathlib, sys, yaml

stack = pathlib.Path(sys.argv[1])
out = pathlib.Path(sys.argv[2])
content_url = sys.argv[3]
content_sha = sys.argv[4]
stack_name = sys.argv[5]
version = sys.argv[6]
content = yaml.safe_load((stack / "content.yaml").read_text())

def sha(p: pathlib.Path) -> str:
    return f"sha256:{hashlib.sha256(p.read_bytes()).hexdigest()}"

artifacts = content.get("artifacts") or {}
vs = content.get("validateSchema") or artifacts.get("validateSchema")
if isinstance(vs, dict):
    vs = vs.get("artifactKey")
schema_version = content.get("schemaVersion", 2)

def materialize_containerfile(body: str) -> dict:
    digest = f"sha256:{hashlib.sha256(body.encode()).hexdigest()}"
    key = f"containerfile:{digest}"
    return {
        "contentRef": {
            "url": f"coin://gp-content/{stack_name}@{version}/{key}",
            "sha256": digest,
        },
        "digest": digest,
        "_artifact_key": key,
        "_body": body,
    }

def materialize_steps(steps: list):
    out = []
    extra_keys = []
    for step in steps or []:
        item = dict(step)
        for block_key in ("run", "build"):
            block = item.get(block_key)
            if not isinstance(block, dict):
                continue
            cf = block.get("containerfile")
            if isinstance(cf, dict) and cf.get("body"):
                mat = materialize_containerfile(cf["body"])
                block = dict(block)
                block["containerfile"] = {
                    "contentRef": mat["contentRef"],
                    "digest": mat["digest"],
                }
                item[block_key] = block
                extra_keys.append((mat["_artifact_key"], mat["_body"]))
        out.append(item)
    return out, extra_keys

stages = []
extra_artifacts = []
for st in content.get("pipeline", {}).get("stages", []):
    item = {"id": st["id"], "name": st["name"]}
    if st.get("when"):
        item["when"] = st["when"]
    if st.get("steps"):
        if schema_version >= 3:
            steps, extras = materialize_steps(st["steps"])
            item["steps"] = steps
            extra_artifacts.extend(extras)
        else:
            item["steps"] = st["steps"]
    stages.append(item)

build = content.get("build") or {}

cref = {
    "parameters": content.get("parameters") or [],
    "pipeline": {"stages": stages},
    "validateSchema": {"artifactKey": vs, "sha256": sha(stack / vs)},
}
if schema_version < 3:
    cref["build"] = build
    cref["deliverables"] = content.get("deliverables") or []
keys = [vs]
containerfiles = []
for cf in artifacts.get("containerfiles") or []:
    key = cf["path"]
    containerfiles.append({"id": cf["id"], "artifactKey": key, "sha256": sha(stack / key)})
    keys.append(key)
if containerfiles:
    cref["artifacts"] = {"containerfiles": containerfiles}
elif content.get("containerfile"):
    cf = content["containerfile"]["artifactKey"]
    cref["containerfile"] = {"artifactKey": cf, "sha256": sha(stack / cf)}
    keys.append(cf)

inline_bodies = {}
for key, body in extra_artifacts:
    keys.append(key)
    inline_bodies[key] = body

controls = content.get("controls") or {}
capabilities = content.get("capabilities") or {}
meta = {
    "url": content_url,
    "sha256": content_sha,
    "buildControls": controls,
    "capabilities": capabilities,
}

(out / "content-ref.json").write_text(json.dumps(cref))
(out / "metadata.json").write_text(json.dumps(meta))
(out / "keys.json").write_text(json.dumps(keys))
(out / "inline-bodies.json").write_text(json.dumps(inline_bodies))
PY

echo "==> upload gp-content artifact bodies to Nexus"
jq -r '.[]' "${PAYLOAD_DIR}/keys.json" | while IFS= read -r key; do
  file="${STACK_DIR}/${key}"
  if [[ ! -f "${file}" ]]; then
    body="$(jq -r --arg k "${key}" '.[$k] // empty' "${PAYLOAD_DIR}/inline-bodies.json")"
    if [[ -n "${body}" ]]; then
      file="${PAYLOAD_DIR}/inline-${key//[:/]/_}"
      printf '%s' "${body}" > "${file}"
    fi
  fi
  if [[ ! -f "${file}" ]]; then
    echo "==> skip missing artifact file for key ${key}"
    continue
  fi
  artifact_url="$(gp_content_artifact_url "${STACK}" "${VERSION}" "${key}")"
  nexus_upload "${artifact_url}" "${file}"
done

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
  echo "==> version already registered in coin-api, updating content_ref"
  cref_json="$(jq -c . "${PAYLOAD_DIR}/content-ref.json")"
  meta_json="$(jq -c . "${PAYLOAD_DIR}/metadata.json")"
  update_code="$(curl -sS -o /dev/null -w '%{http_code}' -X PATCH \
    "${COIN_API_URL}/v1/admin/components/gp-content/${STACK}/versions/${VERSION}" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: ${API_KEY}" \
    -d "{\"metadata\":${meta_json},\"contentRef\":${cref_json}}")"
  if [[ "${update_code}" != "200" ]]; then
    if [[ "${update_code}" == "409" ]]; then
      echo "==> warn: published content_ref is immutable (HTTP 409), keep existing ref"
    elif [[ "${update_code}" == "405" ]]; then
      echo "==> warn: coin-api PATCH not available (HTTP 405), content_ref unchanged"
    else
      echo "coin-api content_ref update failed HTTP ${update_code}" >&2
      exit 1
    fi
  fi
fi
rm -f "${register_tmp}"

jq -r '.[]' "${PAYLOAD_DIR}/keys.json" | while IFS= read -r key; do
  file="${STACK_DIR}/${key}"
  if [[ ! -f "${file}" ]]; then
    body="$(jq -r --arg k "${key}" '.[$k] // empty' "${PAYLOAD_DIR}/inline-bodies.json")"
    if [[ -n "${body}" ]]; then
      file="${PAYLOAD_DIR}/inline-${key//[:/]/_}"
      printf '%s' "${body}" > "${file}"
    fi
  fi
  if [[ ! -f "${file}" ]]; then
    echo "==> skip coin-api artifact ${key} (no file)"
    continue
  fi
  enc_key="$(jq -rn --arg k "${key}" '$k|@uri')"
  echo "==> artifact ${key}"
  put_tmp="$(mktemp)"
  put_code="$(jq -n --rawfile b "${file}" '{body: $b}' | curl -sS -o "${put_tmp}" -w '%{http_code}' -X PUT \
    "${COIN_API_URL}/v1/admin/components/gp-content/${STACK}/versions/${VERSION}/artifacts/${enc_key}" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: ${API_KEY}" \
    -d @-)"
  if [[ "${put_code}" != "200" && "${put_code}" != "201" ]]; then
    echo "==> warn: artifact body ${key} not updated in coin-api (HTTP ${put_code}): $(cat "${put_tmp}")" >&2
  fi
  rm -f "${put_tmp}"
done

rm -rf "${PAYLOAD_DIR}"
echo "==> done gp-content/${STACK}@${VERSION}"
