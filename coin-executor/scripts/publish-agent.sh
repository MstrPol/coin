#!/usr/bin/env bash
# Build coin-agent image from Dockerfile.agent, push to Nexus Docker, register draft in coin-api.
set -euo pipefail

VERSION="${1:?version (e.g. 1.0.0)}"
GOARCH="${2:-amd64}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COIN_API_URL="${COIN_API_URL:-http://localhost:8090}"
API_KEY="${COIN_API_KEY:-dev-local-admin-key}"
GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"

NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"

PUSH_REGISTRY="localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}"
RUNTIME_REGISTRY="nexus:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}"
IMAGE_NAME="coin-agent"
PUSH_REF="${PUSH_REGISTRY}/${IMAGE_NAME}:${VERSION}"
RUNTIME_REF="${RUNTIME_REGISTRY}/${IMAGE_NAME}:${VERSION}"

DOCKER_CONFIG="${DOCKER_CONFIG:-${WORKSPACE:-${ROOT}}/.docker}"
mkdir -p "${DOCKER_CONFIG}"
export DOCKER_CONFIG

for cmd in curl jq docker; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

docker version >/dev/null 2>&1 || {
  echo "docker daemon is not reachable (required for coin-agent build/push)" >&2
  exit 1
}

if ! curl -fsS "${COIN_API_URL}/ready" >/dev/null 2>&1; then
  echo "coin-api is not reachable at ${COIN_API_URL} (required for draft register)" >&2
  exit 1
fi

if [[ -z "${DOCKER_PLATFORM:-}" ]]; then
  case "${GOARCH}" in
    arm64) DOCKER_PLATFORM="linux/arm64" ;;
    amd64) DOCKER_PLATFORM="linux/amd64" ;;
    *) echo "unsupported GOARCH: ${GOARCH}" >&2; exit 1 ;;
  esac
fi

echo "==> build coin-agent ${VERSION} (${GOARCH}) from Dockerfile.agent"
docker build \
  --platform "${DOCKER_PLATFORM}" \
  -f "${ROOT}/Dockerfile.agent" \
  --build-arg VERSION="${VERSION}" \
  --build-arg GOPROXY="${GOPROXY}" \
  -t "${PUSH_REF}" \
  "${ROOT}"

echo "${NEXUS_DOCKER_PASSWORD}" | docker login "localhost:${NEXUS_DOCKER_PORT}" \
  -u "${NEXUS_DOCKER_USER}" --password-stdin
echo "==> push ${PUSH_REF}"
docker push "${PUSH_REF}"

digest=""
if rd="$(docker inspect --format='{{index .RepoDigests 0}}' "${PUSH_REF}" 2>/dev/null || true)"; then
  if [[ "${rd}" == *@sha256:* ]]; then
    digest="${rd#*@}"
  fi
fi

payload="$(jq -n \
  --arg ver "${VERSION}" \
  --arg img "${RUNTIME_REF}" \
  --arg digest "${digest}" \
  '{version: $ver, metadata: {image: $img, digest: $digest, runtime: "coin-agent"}, actor: "publish-agent"}')"

register_tmp="$(mktemp)"
register_code="$(curl -sS -o "${register_tmp}" -w '%{http_code}' -X POST \
  "${COIN_API_URL}/v1/admin/components/agent/coin-agent/versions/drafts" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d "${payload}")"
register_body="$(cat "${register_tmp}")"
if [[ "${register_code}" == "409" ]]; then
  echo "==> draft already exists for agent/coin-agent@${VERSION}, continue"
  echo "    coin-api response: ${register_body}"
elif [[ "${register_code}" != "201" ]]; then
  echo "coin-api register draft failed HTTP ${register_code}: ${register_body}" >&2
  rm -f "${register_tmp}"
  exit 1
fi
rm -f "${register_tmp}"

echo "==> done agent/coin-agent@${VERSION} draft registered image=${RUNTIME_REF} digest=${digest:-<none>}"
echo "    Promote in Platform UI: /platform/runtime/coin-agent → release ${VERSION} → Publish"
if [[ -z "${digest}" ]]; then
  echo "    WARN: digest missing — promote will fail until digest is set in metadata" >&2
fi
