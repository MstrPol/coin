#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-container}"

if [[ "${TARGET}" == "package" ]]; then
  echo "==> coin standard build (maven package)"
  mvn -B -DskipTests package
  exit 0
fi

echo "==> coin standard build (container image)"
DOCKERFILE="${COIN_DOCKERFILE:-}"
if [[ -z "${DOCKERFILE}" || ! -f "${DOCKERFILE}" ]]; then
  echo "Managed Dockerfile not found. Coin must set COIN_DOCKERFILE for container builds." >&2
  exit 1
fi
IMAGE_NAME="${COIN_IMAGE_NAME:-app}"
IMAGE_TAG="${COIN_IMAGE_TAG:-${BUILD_NUMBER:-latest}}"
REGISTRY="${COIN_REGISTRY_PREFIX:-}"
REF="${COIN_IMAGE_REF:-${REGISTRY:+${REGISTRY%/}/}${IMAGE_NAME}:${IMAGE_TAG}}"
mkdir -p .coin
if command -v /kaniko/executor >/dev/null 2>&1; then
  /kaniko/executor --context="${WORKSPACE:-.}" --dockerfile="${DOCKERFILE}" --destination="${REF}" --build-arg="COIN_VERSION=${COIN_VERSION}" --cache=true
elif command -v docker >/dev/null 2>&1; then
  docker build --build-arg "COIN_VERSION=${COIN_VERSION}" -t "${REF}" -f "${DOCKERFILE}" .
else
  echo "No docker or kaniko on agent." >&2
  exit 1
fi
echo "COIN_BUILT_IMAGE=${REF}" > .coin/build.env
