#!/usr/bin/env bash
set -euo pipefail

echo "==> coin standard validate (go-app@1.0.0)"
if [[ ! -f .coin/config.yaml ]]; then
  echo "missing .coin/config.yaml" >&2
  exit 1
fi

if command -v coin-executor >/dev/null 2>&1 && [[ -n "${COIN_MANIFEST_PATH:-}" ]]; then
  coin-executor validate --project .coin/config.yaml --manifest "${COIN_MANIFEST_PATH}"
else
  echo "config file present (full schema check via coin-executor validate in pipeline)"
fi
