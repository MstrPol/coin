#!/usr/bin/env bash
# Build coin-lib Shared Library ZIP and upload to Nexus (no coin-api registry).
set -euo pipefail

VERSION="${1:?version (e.g. 1.0.0)}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${ROOT}/dist"
ZIP="${OUT_DIR}/coin-lib-${VERSION}.zip"

LIB_NAME="coin-lib"
# shellcheck source=lib/maven-url.sh
source "$(dirname "$0")/lib/maven-url.sh"

for cmd in zip curl; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

mkdir -p "${OUT_DIR}"
rm -f "${ZIP}"

(
  cd "${ROOT}"
  zip -qr "${ZIP}" vars src
)

LIB_URL="$(lib_zip_url "${LIB_NAME}" "${VERSION}")"

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

nexus_upload "${LIB_URL}" "${ZIP}"

echo "==> done lib/${LIB_NAME}@${VERSION} -> ${LIB_URL}"
