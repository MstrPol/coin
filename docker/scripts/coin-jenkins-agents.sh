#!/usr/bin/env bash
# coin-jenkins-agents → Gitea coin/coin-jenkins-agents
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

AGENTS="${REPO_ROOT}/coin-jenkins-agents"
[[ -d "${AGENTS}" ]] || { echo "missing ${AGENTS}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-jenkins-agents.git"

echo "==> Gitea: coin/coin-jenkins-agents"
gitea_create_repo "coin-jenkins-agents"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete "${AGENTS}/" "${WORK}/"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-jenkins-agents (local stack)"
git push --force "${url}" main

echo "coin-jenkins-agents ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-jenkins-agents"
