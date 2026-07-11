#!/usr/bin/env bash
# coin-starters → Gitea coin/coin-starters
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

STARTERS="${REPO_ROOT}/coin-starters"
[[ -d "${STARTERS}" ]] || { echo "missing ${STARTERS}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-starters.git"

echo "==> Gitea: coin/coin-starters"
gitea_create_repo "coin-starters"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete "${STARTERS}/" "${WORK}/"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-starters (local stack)"
git push --force "${url}" main

echo "coin-starters ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-starters"
