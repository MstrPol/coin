#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-container}"
GP_SHARED="${COIN_GOLDEN_PATHS_DIR:-coin-golden-paths}/_shared/pack-image.sh"
GRADLE_CMD="gradle"
if [[ -x ./gradlew ]]; then
  GRADLE_CMD="./gradlew"
fi

if [[ "${TARGET}" == "package" ]]; then
  echo "==> coin standard build (gradle package)"
  ${GRADLE_CMD} build -x test --no-daemon
  exit 0
fi

echo "==> coin standard build (gradle package, native)"
${GRADLE_CMD} build -x test --no-daemon

echo "==> coin standard build (container image, pack)"
# shellcheck disable=SC1090
source "${GP_SHARED}"
coin_pack_image
