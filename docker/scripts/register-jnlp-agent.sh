#!/usr/bin/env bash
# Mirror Jenkins inbound-agent → Nexus coin-docker + register agent/jnlp в coin-api.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

UPSTREAM="${COIN_JNLP_UPSTREAM:-jenkins/inbound-agent:3327.v868139a_d00e0-8}"
VERSION="${COIN_JNLP_VERSION:-${UPSTREAM#*:}}"
JNLP_REPO="ci-jnlp"

NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"

PUSH_REGISTRY="localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}"
RUNTIME_REGISTRY="nexus:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}"
PUSH_REF="${PUSH_REGISTRY}/${JNLP_REPO}:${VERSION}"
RUNTIME_REF="${RUNTIME_REGISTRY}/${JNLP_REPO}:${VERSION}"

COIN_API_URL="${COIN_API_URL:-http://localhost:8090}"
API_KEY="${COIN_PUBLISHER_API_KEY:-dev-local-publisher-key}"

for cmd in curl jq docker; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "missing required command: ${cmd}" >&2; exit 1; }
done

if [[ -z "${DOCKER_PLATFORM:-}" ]]; then
  case "$(uname -m)" in
    arm64|aarch64) DOCKER_PLATFORM="linux/arm64" ;;
    x86_64|amd64) DOCKER_PLATFORM="linux/amd64" ;;
    *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
  esac
fi

nexus_manifest_code() {
  curl -s -o /dev/null -w '%{http_code}' \
    -u "${NEXUS_DOCKER_USER}:${NEXUS_DOCKER_PASSWORD}" \
    -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    "http://localhost:${NEXUS_DOCKER_PORT}/v2/${NEXUS_DOCKER_REPO}/${JNLP_REPO}/manifests/${VERSION}" \
    || true
}

push_jnlp_image() {
  local code
  code="$(nexus_manifest_code)"
  if [[ "${code}" == "200" ]]; then
    echo "==> already in Nexus, skip push: ${PUSH_REF}"
    return 0
  fi

  echo "==> pull ${UPSTREAM} (${DOCKER_PLATFORM})"
  docker pull --platform "${DOCKER_PLATFORM}" "${UPSTREAM}"
  docker tag "${UPSTREAM}" "${PUSH_REF}"

  echo "==> push ${PUSH_REF}"
  echo "${NEXUS_DOCKER_PASSWORD}" | docker login "localhost:${NEXUS_DOCKER_PORT}" \
    -u "${NEXUS_DOCKER_USER}" --password-stdin
  docker push "${PUSH_REF}"
}

image_digest() {
  local rd=""
  if docker image inspect "${PUSH_REF}" >/dev/null 2>&1; then
    rd="$(docker inspect --format='{{index .RepoDigests 0}}' "${PUSH_REF}" 2>/dev/null || true)"
  else
    docker pull "${PUSH_REF}" >/dev/null 2>&1 || true
    rd="$(docker inspect --format='{{index .RepoDigests 0}}' "${PUSH_REF}" 2>/dev/null || true)"
  fi
  if [[ "${rd}" == *@sha256:* ]]; then
    echo "${rd#*@}"
  fi
}

register_jnlp() {
  local digest="$1"
  local payload
  if [[ -n "${digest}" ]]; then
    payload="$(jq -n \
      --arg ver "${VERSION}" \
      --arg img "${RUNTIME_REF}" \
      --arg digest "${digest}" \
      --arg upstream "${UPSTREAM}" \
      '{version: $ver, metadata: {image: $img, digest: $digest, runtime: "jnlp", stack: "jnlp", upstream: $upstream}, actor: "register-jnlp-agent"}')"
  else
    payload="$(jq -n \
      --arg ver "${VERSION}" \
      --arg img "${RUNTIME_REF}" \
      --arg upstream "${UPSTREAM}" \
      '{version: $ver, metadata: {image: $img, runtime: "jnlp", stack: "jnlp", upstream: $upstream}, actor: "register-jnlp-agent"}')"
  fi

  local register_tmp register_code
  register_tmp="$(mktemp)"
  register_code="$(curl -sS -o "${register_tmp}" -w '%{http_code}' -X POST \
    "${COIN_API_URL}/v1/admin/components/agent/jnlp/versions" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: ${API_KEY}" \
    -d "${payload}")"
  if [[ "${register_code}" != "201" && "${register_code}" != "409" ]]; then
    echo "coin-api register failed HTTP ${register_code}: $(cat "${register_tmp}")" >&2
    rm -f "${register_tmp}"
    exit 1
  fi
  if [[ "${register_code}" == "409" ]]; then
    echo "==> agent/jnlp@${VERSION} already registered in coin-api"
    local detail_tmp detail_code registered_image
    detail_tmp="$(mktemp)"
    detail_code="$(curl -sS -o "${detail_tmp}" -w '%{http_code}' \
      -H "X-API-Key: ${API_KEY}" \
      "${COIN_API_URL}/v1/admin/components/agent/jnlp/versions/${VERSION}")"
    if [[ "${detail_code}" == "200" ]]; then
      registered_image="$(jq -r '.metadata.image // empty' "${detail_tmp}")"
      if [[ -n "${registered_image}" && "${registered_image}" != "${RUNTIME_REF}" ]]; then
        echo "WARN: в coin-api image=${registered_image}, ожидается ${RUNTIME_REF}" >&2
        echo "      смените COIN_JNLP_VERSION или удалите версию в registry" >&2
      fi
    fi
    rm -f "${detail_tmp}"
  else
    echo "==> registered agent/jnlp@${VERSION} image=${RUNTIME_REF}"
  fi
  rm -f "${register_tmp}"
}

push_jnlp_image
digest="$(image_digest)"
echo "digest=${digest:-<none>}"
register_jnlp "${digest}"
