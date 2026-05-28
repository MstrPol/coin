#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-package}"

case "${TARGET}" in
  package)
    echo "==> coin standard build (python package)"
    uv build
    ;;
  container)
    echo "==> coin standard build (container image)"
    DOCKERFILE="${COIN_DOCKERFILE:-}"
    if [[ -z "${DOCKERFILE}" || ! -f "${DOCKERFILE}" ]]; then
      echo "Managed Dockerfile not found. Coin must set COIN_DOCKERFILE for container builds." >&2
      exit 1
    fi

    IMAGE_NAME="${COIN_IMAGE_NAME:-app}"
    IMAGE_TAG="${COIN_IMAGE_TAG:-${BUILD_NUMBER:-latest}}"
    REGISTRY="${COIN_REGISTRY_PREFIX:-}"
    REF="${COIN_IMAGE_REF:-}"

    if [[ -z "${REF}" ]]; then
      if [[ -n "${REGISTRY}" ]]; then
        REF="${REGISTRY%/}/${IMAGE_NAME}:${IMAGE_TAG}"
      else
        REF="${IMAGE_NAME}:${IMAGE_TAG}"
      fi
    fi

    echo "Building image: ${REF}"
    mkdir -p .coin

    if command -v /kaniko/executor >/dev/null 2>&1; then
      /kaniko/executor \
        --context="${WORKSPACE:-.}" \
        --dockerfile="${DOCKERFILE}" \
        --destination="${REF}" \
        --build-arg="COIN_VERSION=${COIN_VERSION}" \
        --cache=true
    elif command -v docker >/dev/null 2>&1; then
      docker build --build-arg "COIN_VERSION=${COIN_VERSION}" -t "${REF}" -f "${DOCKERFILE}" .
    else
      echo "No docker or kaniko on agent." >&2
      exit 1
    fi

    echo "COIN_BUILT_IMAGE=${REF}" > .coin/build.env
    ;;
  *)
    echo "Unknown COIN_BUILD_TARGET=${TARGET}" >&2
    exit 1
    ;;
esac
