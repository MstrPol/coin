#!/usr/bin/env bash
# Upload coin-executor binary to Nexus (for agent image bake). No coin-api registry.
set -euo pipefail

VERSION="${1:?version (e.g. 1.0.0)}"
GOARCH="${2:?goarch (amd64|arm64)}"
BINARY="${3:-coin-executor}"

# shellcheck source=lib/maven-url.sh
source "$(dirname "$0")/lib/maven-url.sh"

for cmd in curl; do
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

echo "==> done Nexus upload coin-executor@${VERSION} (${GOARCH}) at ${REMOTE}"
