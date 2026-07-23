#!/usr/bin/env bash
# Сборка coin-executor/Dockerfile и push в локальный Nexus (dev-stand).
# Каждый запуск: полная пересборка (--no-cache) + push с перезаписью тега.
# Базы — локальные docker-image:// context'ы (без resolve metadata на Hub).
set -euo pipefail

stand_base="$(cd "$(dirname "$0")/../../dev-stand-base" && pwd)"
coin_executor="$(cd "$(dirname "$0")/../../../coin-executor" && pwd)"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

VERSION="${VERSION:-1.0.0}"
PLATFORM="${DOCKER_PLATFORM:-linux/arm64}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"
GOLANG_REF="${GOLANG_REF:-golang:1.24-bookworm}"
BUILDKIT_REF="${BUILDKIT_REF:-moby/buildkit:v0.21.0-rootless}"
# PULL_BASE=1 — docker pull баз с Hub перед сборкой (когда DNS/Hub доступны)
PULL_BASE="${PULL_BASE:-0}"

registry="localhost:${NEXUS_DOCKER_PORT}"
image="${registry}/${NEXUS_DOCKER_REPO}/coin-agent:${VERSION}"

[[ -f "${coin_executor}/Dockerfile" ]] || {
  echo "нет Dockerfile: ${coin_executor}/Dockerfile" >&2
  exit 1
}

ensure_base() {
  local ref=$1
  if [[ "${PULL_BASE}" == "1" ]]; then
    echo "==> pull base ${ref}"
    docker pull --platform "${PLATFORM}" "${ref}"
  fi
  if ! docker image inspect "${ref}" >/dev/null 2>&1; then
    echo "нет локального образа: ${ref}" >&2
    echo "  один раз при доступе к Hub: docker pull --platform ${PLATFORM} ${ref}" >&2
    echo "  или: PULL_BASE=1 make publish-executor" >&2
    exit 1
  fi
}

ensure_base "${GOLANG_REF}"
ensure_base "${BUILDKIT_REF}"

echo "==> bases (docker-image:// local contexts)"
echo "    golang   ← ${GOLANG_REF}"
echo "    buildkit ← ${BUILDKIT_REF}"

echo "==> build ${image} (${PLATFORM}, no-cache)"
DOCKER_BUILDKIT=1 docker build \
  --platform "${PLATFORM}" \
  --no-cache \
  --build-context "golang=docker-image://${GOLANG_REF}" \
  --build-context "buildkit=docker-image://${BUILDKIT_REF}" \
  --build-arg "VERSION=${VERSION}" \
  -t "${image}" \
  -f "${coin_executor}/Dockerfile" \
  "${coin_executor}"

echo "==> login ${registry}"
echo "${NEXUS_DOCKER_PASSWORD}" | docker login "${registry}" \
  -u "${NEXUS_DOCKER_USER}" --password-stdin

echo "==> push ${image} (overwrite tag)"
docker push "${image}"

echo "OK: ${image}"
