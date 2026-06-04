#!/usr/bin/env bash
# coin-platform → Gitea coin/coin-platform (golden-paths, starters, agents).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

PLATFORM="${REPO_ROOT}/coin-platform"
[[ -d "${PLATFORM}" ]] || { echo "missing ${PLATFORM}" >&2; exit 1; }

echo "==> Gitea: coin/coin-platform"
gitea_create_repo "coin-platform"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete "${PLATFORM}/" "${WORK}/"

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-platform.git"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-platform (local stack)"
git push --force "${url}" main

echo "coin-platform ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-platform"
