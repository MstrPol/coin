#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-package}"
GP_ROOT="${COIN_GOLDEN_PATHS_DIR:-${COIN_PLATFORM_DIR:+${COIN_PLATFORM_DIR}/golden-paths}}"
GP_SHARED="${GP_ROOT}/_shared/pack-image.sh"

case "${TARGET}" in
  package)
    echo "==> coin standard build (python-pip package)"
    python -m pip install --upgrade build
    python -m build
    ;;
  container)
    echo "==> coin standard build (python venv, native)"
    python -m venv .venv
    # shellcheck disable=SC1091
    source .venv/bin/activate
    python -m pip install --upgrade pip
    python -m pip install -r requirements.txt

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
