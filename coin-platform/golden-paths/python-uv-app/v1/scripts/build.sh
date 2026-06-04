#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-package}"
GP_ROOT="${COIN_GOLDEN_PATHS_DIR:-${COIN_PLATFORM_DIR:+${COIN_PLATFORM_DIR}/golden-paths}}"
GP_SHARED="${GP_ROOT}/_shared/pack-image.sh"

case "${TARGET}" in
  package)
    echo "==> coin standard build (python package)"
    uv build
    ;;
  container)
    echo "==> coin standard build (python deps, native)"
    uv sync --frozen --no-install-project --no-dev
    uv sync --frozen --no-dev

    echo "==> coin standard build (container image, pack)"
    # shellcheck disable=SC1090
    source "${GP_SHARED}"
    coin_pack_image
    ;;
  *)
    echo "Unknown COIN_BUILD_TARGET=${TARGET}" >&2
    exit 1
    ;;
esac
