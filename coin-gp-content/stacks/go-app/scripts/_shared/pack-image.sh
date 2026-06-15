#!/usr/bin/env bash
# Shared pack helper: runtime-only Dockerfile + dist/ → registry image (docker; kaniko later).
set -euo pipefail

coin_pack_image() {
  local dockerfile="${COIN_DOCKERFILE:-.coin/Dockerfile}"
  local registry="${COIN_REGISTRY_PREFIX:-localhost:8082/coin-docker}"
  local project
  project="$(grep -E '^  name:' .coin/config.yaml 2>/dev/null | awk '{print $2}' || true)"
  project="${project:-app}"
  local gp_version
  gp_version="$(python3 -c 'import json; print(json.load(open(".coin/manifest.json"))["goldenPath"]["version"])' 2>/dev/null || echo local)"
  local tag="${COIN_IMAGE_TAG:-${gp_version}}"
  local rendered=".coin/Dockerfile.rendered"

  sed \
    -e 's/{{APP_PORT}}/8080/g' \
    -e 's|{{APP_CMD}}|["/app/app"]|g' \
    -e "s/{{COIN_VERSION}}/${tag}/g" \
    "${dockerfile}" > "${rendered}"

  local image_ref="${registry%/}/${project}:${tag}"
  echo "==> docker build ${image_ref}"
  docker build -f "${rendered}" -t "${image_ref}" .

  mkdir -p .coin
  {
    echo "COIN_BUILT_IMAGE=${image_ref}"
    echo "COIN_IMAGE_TAG=${tag}"
  } >> .coin/build.env
}
