#!/usr/bin/env bash
# E2E: demo-go-app (buildkit), demo-go-app-bp (buildpack), demo-go-app-df (dockerfile).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

PUBLISH="${COIN_E2E_PUBLISH:-false}"
TIMEOUT_SEC="${COIN_E2E_TIMEOUT_SEC:-1200}"
JOBS=(demo-go-app demo-go-app-bp demo-go-app-df)
if [[ -n "${COIN_E2E_JOBS:-}" ]]; then
  # shellcheck disable=SC2206
  JOBS=(${COIN_E2E_JOBS})
fi

if [[ "${COIN_E2E_SKIP_PRUNE:-0}" != "1" ]]; then
  COIN_PRUNE_DOCKER=1 bash "${ROOT}/scripts/prune-k3s-disk.sh" --all
fi

failed=()
for job in "${JOBS[@]}"; do
  echo ""
  echo "════════════════════════════════════════════════════"
  echo "E2E ${job} (publish=${PUBLISH})"
  echo "════════════════════════════════════════════════════"
  if COIN_E2E_JOB="${job}" COIN_E2E_SKIP_PRUNE=1 COIN_E2E_PUBLISH="${PUBLISH}" \
    COIN_E2E_TIMEOUT_SEC="${TIMEOUT_SEC}" \
    bash "${ROOT}/scripts/e2e-demo-go-app.sh"; then
    echo "OK: ${job}"
  else
    echo "FAIL: ${job}" >&2
    failed+=("${job}")
  fi
done

echo ""
if [[ ${#failed[@]} -eq 0 ]]; then
  echo "OK: all build-engine E2E jobs SUCCESS (${JOBS[*]})"
  exit 0
fi
echo "FAIL: ${#failed[@]} job(s): ${failed[*]}" >&2
exit 1
