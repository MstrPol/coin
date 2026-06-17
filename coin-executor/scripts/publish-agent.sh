#!/usr/bin/env bash
# Build coin-agent image, push to Nexus Docker, register agent/coin-agent in coin-api.
set -euo pipefail

VERSION="${1:?version (e.g. 1.0.0)}"
GOARCH="${2:-amd64}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COIN_API_URL="${COIN_API_URL:-http://localhost:8090}"
API_KEY="${COIN_API_KEY:-dev-local-admin-key}"

NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"

PUSH_REGISTRY="localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}"
RUNTIME_REGISTRY="nexus:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}"
IMAGE_NAME="coin-agent"
PUSH_REF="${PUSH_REGISTRY}/${IMAGE_NAME}:${VERSION}"
RUNTIME_REF="${RUNTIME_REGISTRY}/${IMAGE_NAME}:${VERSION}"

for cmd in curl jq docker; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

if [[ -z "${DOCKER_PLATFORM:-}" ]]; then
  case "${GOARCH}" in
    arm64) DOCKER_PLATFORM="linux/arm64" ;;
    amd64) DOCKER_PLATFORM="linux/amd64" ;;
    *) echo "unsupported GOARCH: ${GOARCH}" >&2; exit 1 ;;
  esac
fi

binary="${COIN_EXECUTOR_BINARY:-${ROOT}/coin-executor}"
if [[ ! -f "${binary}" ]]; then
  echo "coin-executor binary not found: ${binary}" >&2
  echo "build first: GOOS=linux GOARCH=${GOARCH} go build -o coin-executor ./cmd/coin-executor" >&2
  exit 1
fi

build_ctx="$(mktemp -d)"
trap 'rm -rf "${build_ctx}"' EXIT
cp "${binary}" "${build_ctx}/coin-executor"
cp "${ROOT}/Dockerfile.agent" "${build_ctx}/Dockerfile"
cp "${ROOT}/buildkitd.toml" "${build_ctx}/buildkitd.toml"
cp "${ROOT}/podman-containers.conf" "${build_ctx}/podman-containers.conf"
cp "${ROOT}/podman-storage.conf" "${build_ctx}/podman-storage.conf"
cp "${ROOT}/podman-registries.conf" "${build_ctx}/podman-registries.conf"

PACK_VERSION="0.36.2"
pack_bin="${COIN_PACK_BINARY:-}"
if [[ -z "${pack_bin}" && -f "${ROOT}/pack.${GOARCH}" ]]; then
  pack_bin="${ROOT}/pack.${GOARCH}"
fi
if [[ -z "${pack_bin}" ]]; then
  pack_tgz="${build_ctx}/pack.tgz"
  pack_url="https://github.com/buildpacks/pack/releases/download/v${PACK_VERSION}/pack-v${PACK_VERSION}-linux-${GOARCH}.tgz"
  echo "==> download pack ${PACK_VERSION} (${GOARCH})"
  if curl -fsSL --retry 5 --retry-delay 3 --connect-timeout 30 "${pack_url}" -o "${pack_tgz}"; then
    tar -xzf "${pack_tgz}" -C "${build_ctx}" pack
    pack_bin="${build_ctx}/pack"
  elif docker image inspect "${PUSH_REF}" >/dev/null 2>&1; then
    echo "==> reuse pack from existing ${PUSH_REF}"
    cid="$(docker create "${PUSH_REF}")"
  elif docker image inspect "${PUSH_REGISTRY}/${IMAGE_NAME}:latest" >/dev/null 2>&1; then
    echo "==> reuse pack from existing ${PUSH_REGISTRY}/${IMAGE_NAME}:latest"
    cid="$(docker create "${PUSH_REGISTRY}/${IMAGE_NAME}:latest")"
  else
    echo "failed to download pack and no local agent image to reuse" >&2
    exit 1
  fi
fi
if [[ -n "${cid:-}" ]]; then
  docker cp "${cid}:/usr/local/bin/pack" "${build_ctx}/pack"
  docker rm "${cid}" >/dev/null
  pack_bin="${build_ctx}/pack"
fi
cp "${pack_bin}" "${build_ctx}/pack"
chmod +x "${build_ctx}/pack"

BUILDER_LOCAL="localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}/paketo-builder-jammy-base:latest"
BUILDER_RUNTIME="nexus:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}/paketo-builder-jammy-base:latest"
if docker image inspect "${BUILDER_LOCAL}" >/dev/null 2>&1; then
  echo "==> bake buildpack builder into agent: ${BUILDER_RUNTIME}"
  docker tag "${BUILDER_LOCAL}" "${BUILDER_RUNTIME}"
  docker save "${BUILDER_RUNTIME}" -o "${build_ctx}/paketo-builder.tar"
else
  echo "missing builder image ${BUILDER_LOCAL}; run docker/scripts/mirror-paketo-builder.sh first" >&2
  exit 1
fi

echo "==> build coin-agent ${VERSION} (${GOARCH})"
docker build \
  --platform "${DOCKER_PLATFORM}" \
  -f "${build_ctx}/Dockerfile" \
  -t "${PUSH_REF}" \
  "${build_ctx}"

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
  --arg arch "${GOARCH}" \
  '{version: $ver, metadata: {image: $img, digest: $digest, runtime: "coin-agent", goarch: $arch}, actor: "publish-agent"}')"

register_tmp="$(mktemp)"
register_code="$(curl -sS -o "${register_tmp}" -w '%{http_code}' -X POST \
  "${COIN_API_URL}/v1/admin/components/agent/coin-agent/versions" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d "${payload}")"
if [[ "${register_code}" != "201" && "${register_code}" != "409" ]]; then
  echo "coin-api register failed HTTP ${register_code}: $(cat "${register_tmp}")" >&2
  rm -f "${register_tmp}"
  exit 1
fi
rm -f "${register_tmp}"

echo "==> done agent/coin-agent@${VERSION} image=${RUNTIME_REF} digest=${digest:-<none>}"
