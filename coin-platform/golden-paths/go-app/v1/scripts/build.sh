#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-container}"
GP_ROOT="${COIN_GOLDEN_PATHS_DIR:-${COIN_PLATFORM_DIR:+${COIN_PLATFORM_DIR}/golden-paths}}"
GP_SHARED="${GP_ROOT}/_shared/pack-image.sh"

case "${TARGET}" in
  package)
    echo "==> coin standard build (go binary)"
    mkdir -p dist
    go build -trimpath -o dist/app .
    ;;
  container)
    echo "==> coin standard build (go binary, native)"
    mkdir -p dist
    go build -trimpath -o dist/app .

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
