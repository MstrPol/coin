#!/usr/bin/env bash
set -euo pipefail

TARGET="${COIN_BUILD_TARGET:-package}"

case "${TARGET}" in
  package)
    echo "==> coin standard publish (PyPI / Nexus)"
    python -m pip install --upgrade twine
    python -m twine upload dist/*
    ;;
  container)
    echo "==> coin standard publish (container registry)"
    if [[ -f .coin/build.env ]]; then
      # shellcheck source=/dev/null
      source .coin/build.env
    fi
    if [[ -z "${COIN_BUILT_IMAGE:-}" ]]; then
      echo "COIN_BUILT_IMAGE not set. Run build with target=container first." >&2
      exit 1
    fi
    if command -v /kaniko/executor >/dev/null 2>&1; then
      echo "Image pushed during kaniko build: ${COIN_BUILT_IMAGE}"
    elif command -v docker >/dev/null 2>&1; then
      docker push "${COIN_BUILT_IMAGE}"
    else
      echo "Nothing to push or missing docker." >&2
      exit 1
    fi
    ;;
esac
