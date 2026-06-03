#!/usr/bin/env bash
set -euo pipefail

echo "==> coin standard publish (java-maven/container)"
if [[ -f .coin/build.env ]]; then
  # shellcheck source=/dev/null
  source .coin/build.env
fi
if [[ -n "${COIN_BUILT_IMAGE:-}" ]]; then
  if command -v /kaniko/executor >/dev/null 2>&1; then
    echo "Image pushed during kaniko build: ${COIN_BUILT_IMAGE}"
  elif command -v docker >/dev/null 2>&1; then
    docker push "${COIN_BUILT_IMAGE}"
  fi
else
  mvn -B deploy
fi
