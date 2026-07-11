#!/usr/bin/env bash
# Полный bootstrap build-engine E2E после wipe-e2e-fresh.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

VERSION="${COIN_E2E_VERSION:-1.0.0}"
GOARCH="${GOARCH:-amd64}"
API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_PUBLISHER_API_KEY:-dev-local-admin-key}"

export COIN_E2E_VERSION="${VERSION}"
export COIN_API_URL="${API}"
export COIN_API_KEY="${KEY}"
export NEXUS_URL="${NEXUS_URL:-http://localhost:8081}"
export NEXUS_USER="${NEXUS_USER:-admin}"
export NEXUS_PASSWORD="${NEXUS_ADMIN_PASSWORD:-coin12345}"

echo "==> [1/7] coin-api ready"
curl -fsS "${API}/ready" >/dev/null

echo "==> [2/7] deploy coin-lib"
make -C "${ROOT}" coin-lib

echo "==> [3/7] upload coin-executor binary to Nexus ${VERSION}"
(
  cd "${REPO_ROOT}/coin-executor"
  GOOS=linux GOARCH="${GOARCH}" CGO_ENABLED=0 go build -trimpath \
    -ldflags "-X main.Version=${VERSION}" -o coin-executor ./cmd/coin-executor
  chmod +x scripts/publish-executor.sh
  ./scripts/publish-executor.sh "${VERSION}" "${GOARCH}" coin-executor
)

echo "==> [4/7] publish coin-agent/${VERSION}"
chmod +x "${REPO_ROOT}/coin-executor/scripts/publish-agent.sh"
VERSION="${VERSION}" GOARCH="${GOARCH}" "${REPO_ROOT}/coin-executor/scripts/publish-agent.sh" "${VERSION}" "${GOARCH}"

echo "==> [4b/7] promote coin-agent/${VERSION} (manual gate)"
api_post() {
  local path="$1" body="$2"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X POST "${API}${path}" \
    -H "X-API-Key: ${KEY}" -H "Content-Type: application/json" -d "${body}")"
  if [[ "${code}" != "201" && "${code}" != "200" && "${code}" != "409" ]]; then
    echo "POST ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  rm -f "${tmp}"
}
api_post "/v1/admin/components/agent/coin-agent/versions/${VERSION}/promote" '{"actor":"e2e-bootstrap"}'

echo "==> [5/7] seed GP stack (lib, gp-content, go-app@${VERSION})"
"${ROOT}/scripts/seed-jenkins-lib-stack.sh"

echo "==> [6/7] samples (demo-go-app Jenkins job)"
make -C "${ROOT}" samples

echo "==> [7/7] API E2E checks"
"${ROOT}/scripts/e2e-jenkins-lib.sh"

echo "bootstrap OK — run: cd docker && make e2e-demo-go-app"
