#!/usr/bin/env bash
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

TARGET="${COIN_BUILD_TARGET:-container}"

case "${TARGET}" in
  package)
    echo "==> coin standard build (go binary)"
    mkdir -p dist
    go build -buildvcs=false -trimpath -o dist/app .
    ;;
  container)
    echo "==> coin standard build (go binary, native)"
    mkdir -p dist
    go build -buildvcs=false -trimpath -o dist/app .

    echo "==> coin standard build (container image, pack)"
    coin_pack_image
    python3 - "${COIN_BUILT_IMAGE:-}" <<'PY'
import json, os, sys
ref = sys.argv[1] if len(sys.argv) > 1 else os.environ.get("COIN_BUILT_IMAGE", "")
if not ref:
    raise SystemExit(0)
path = ".coin/outputs.json"
items = []
if os.path.exists(path):
    with open(path) as f:
        items = json.load(f)
items = [o for o in items if o.get("name") != "app"]
items.append({"name": "app", "type": "image", "ref": ref})
with open(path, "w") as f:
    json.dump(items, f)
PY
    ;;
  *)
    echo "Unknown COIN_BUILD_TARGET=${TARGET}" >&2
    exit 1
    ;;
esac
