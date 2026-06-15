#!/usr/bin/env bash
# Upload coin-executor binary to Nexus and register in coin-api.
set -euo pipefail

VERSION="${1:?version (e.g. 1.0.0)}"
GOARCH="${2:?goarch (amd64|arm64)}"
BINARY="${3:-coin-executor}"

COIN_API_URL="${COIN_API_URL:-http://coin-api:8090}"
API_KEY="${COIN_API_KEY:-dev-local-admin-key}"
# shellcheck source=lib/maven-url.sh
source "$(dirname "$0")/lib/maven-url.sh"

for cmd in curl jq; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

if [[ ! -f "${BINARY}" ]]; then
  echo "binary not found: ${BINARY}" >&2
  exit 1
fi

REMOTE="$(executor_binary_url "${VERSION}" "${GOARCH}")"

nexus_upload() {
  local url="$1" file="$2"
  local auth="${NEXUS_USER:-admin}:${NEXUS_PASSWORD:-coin12345}"
  local code
  code="$(curl -s -o /dev/null -w '%{http_code}' -u "${auth}" -I "${url}" || true)"
  if [[ "${code}" == "200" ]]; then
    echo "==> already in Nexus, skip upload: ${url}"
    return 0
  fi
  echo "==> upload ${file} -> ${url}"
  local body tmp
  tmp="$(mktemp)"
  code="$(curl -sS -u "${auth}" -o "${tmp}" -w '%{http_code}' \
    -H 'Content-Type: application/octet-stream' \
    --upload-file "${file}" "${url}")"
  if [[ "${code}" == "201" || "${code}" == "200" || "${code}" == "204" ]]; then
    rm -f "${tmp}"
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

nexus_upload "${REMOTE}" "${BINARY}"

SHA256="sha256:$(sha256sum "${BINARY}" | awk '{print $1}')"

echo "==> register executor/coin-executor@${VERSION}"
payload="$(jq -n \
  --arg ver "${VERSION}" \
  --arg url "${REMOTE}" \
  --arg arch "${GOARCH}" \
  --arg sha "${SHA256}" \
  '{version: $ver, metadata: {url: $url, goarch: $arch, sha256: $sha}, actor: "jenkins-executor"}')"
register_tmp="$(mktemp)"
register_code="$(curl -sS -o "${register_tmp}" -w '%{http_code}' -X POST \
  "${COIN_API_URL}/v1/admin/components/executor/coin-executor/versions" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d "${payload}")"
if [[ "${register_code}" != "201" && "${register_code}" != "409" ]]; then
  echo "coin-api register failed HTTP ${register_code}: $(cat "${register_tmp}")" >&2
  rm -f "${register_tmp}"
  exit 1
fi
if [[ "${register_code}" == "409" ]]; then
  echo "==> version already registered in coin-api"
fi
rm -f "${register_tmp}"

echo "==> done executor/coin-executor@${VERSION} (${GOARCH})"
