#!/usr/bin/env bash
# coin-platform → Gitea coin/coin-platform (golden-paths, starters, agents).
# Перед push: из Gitea подтягиваются только CI-файлы, не изменённые локально с прошлого push.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

PLATFORM="${REPO_ROOT}/coin-platform"
[[ -d "${PLATFORM}" ]] || { echo "missing ${PLATFORM}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-platform.git"
SYNC_STATE="${ROOT}/.coin-platform-sync.sha256"

# Jenkins agents-build пишет в Gitea — merge по правилу «локальное важнее».
GITEA_SYNC_PATHS=(
  agents/catalog.yaml
)

file_sha256() {
  local f="$1"
  [[ -f "${f}" ]] || return 0
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${f}" | awk '{print $1}'
  else
    shasum -a 256 "${f}" | awk '{print $1}'
  fi
}

stored_hash() {
  local rel="$1"
  [[ -f "${SYNC_STATE}" ]] || return 0
  awk -F '\t' -v p="${rel}" '$1 == p { print $2; exit }' "${SYNC_STATE}"
}

save_sync_state() {
  local rel hash tmp
  tmp="$(mktemp)"
  for rel in "${GITEA_SYNC_PATHS[@]}"; do
    hash="$(file_sha256 "${PLATFORM}/${rel}")"
    [[ -n "${hash}" ]] && printf '%s\t%s\n' "${rel}" "${hash}" >> "${tmp}"
  done
  mv "${tmp}" "${SYNC_STATE}"
}

sync_platform_from_gitea() {
  local clone_dir rel local_file remote_file
  local local_hash baseline pulled=0 kept=0
  clone_dir="$(mktemp -d)"

  if ! git clone --quiet --depth 1 -b main "${url}" "${clone_dir}" 2>/dev/null; then
    rm -rf "${clone_dir}"
    return 0
  fi

  for rel in "${GITEA_SYNC_PATHS[@]}"; do
    remote_file="${clone_dir}/${rel}"
    local_file="${PLATFORM}/${rel}"
    [[ -f "${remote_file}" ]] || continue

    local_hash="$(file_sha256 "${local_file}")"
    baseline="$(stored_hash "${rel}")"

    if [[ -n "${baseline}" && "${local_hash}" != "${baseline}" ]]; then
      echo "    · ${rel} (local edits, keep)"
      kept=1
      continue
    fi

    mkdir -p "${PLATFORM}/$(dirname "${rel}")"
    cp "${remote_file}" "${local_file}"
    echo "    ← ${rel}"
    pulled=1
  done
  rm -rf "${clone_dir}"

  if [[ "${pulled}" -eq 1 ]]; then
    echo "==> synced from Gitea → coin-platform/"
  elif [[ "${kept}" -eq 1 ]]; then
    echo "==> Gitea sync skipped (local changes in CI paths)"
  fi
}

echo "==> Gitea: coin/coin-platform"
gitea_create_repo "coin-platform"

sync_platform_from_gitea

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete "${PLATFORM}/" "${WORK}/"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-platform (local stack)"
git push --force "${url}" main

save_sync_state

echo "coin-platform ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-platform"
