#!/usr/bin/env bash
# Shared helper: упаковка pre-built артефактов в OCI image (native build — в build.sh GP).
# Usage: source .../pack-image.sh && coin_pack_image

coin_pack_image() {
  local dockerfile="${COIN_DOCKERFILE:-}"
  if [[ -z "${dockerfile}" || ! -f "${dockerfile}" ]]; then
    echo "Managed Dockerfile not found. Coin must set COIN_DOCKERFILE for container builds." >&2
    exit 1
  fi

  local image_name="${COIN_IMAGE_NAME:-app}"
  local image_tag="${COIN_IMAGE_TAG:-${BUILD_NUMBER:-latest}}"
  local registry="${COIN_REGISTRY_PREFIX:-}"
  local ref="${COIN_IMAGE_REF:-}"
  local coin_version="${COIN_VERSION:-0.0.0-local}"

  if [[ -z "${ref}" ]]; then
    if [[ -n "${registry}" ]]; then
      ref="${registry%/}/${image_name}:${image_tag}"
    else
      ref="${image_name}:${image_tag}"
    fi
  fi

  echo "==> pack container image: ${ref}"
  mkdir -p .coin

  if command -v /kaniko/executor >/dev/null 2>&1; then
    /kaniko/executor \
      --context="${WORKSPACE:-.}" \
      --dockerfile="${dockerfile}" \
      --destination="${ref}" \
      --build-arg="COIN_VERSION=${coin_version}" \
      --cache=true
  elif command -v docker >/dev/null 2>&1; then
    docker build \
      --build-arg "COIN_VERSION=${coin_version}" \
      -t "${ref}" \
      -f "${dockerfile}" \
      .
  else
    echo "No docker or kaniko on agent." >&2
    exit 1
  fi

  echo "COIN_BUILT_IMAGE=${ref}" > .coin/build.env
}
