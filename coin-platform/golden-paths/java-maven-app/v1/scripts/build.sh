#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-container}"
GP_ROOT="${COIN_GOLDEN_PATHS_DIR:-${COIN_PLATFORM_DIR:+${COIN_PLATFORM_DIR}/golden-paths}}"
GP_SHARED="${GP_ROOT}/_shared/pack-image.sh"

if [[ "${TARGET}" == "package" ]]; then
  echo "==> coin standard build (maven package)"
  mvn -B -DskipTests package
  exit 0
fi

echo "==> coin standard build (maven package, native)"
mvn -B -DskipTests package

echo "==> coin standard build (container image, pack)"
# shellcheck disable=SC1090
source "${GP_SHARED}"
coin_pack_image
