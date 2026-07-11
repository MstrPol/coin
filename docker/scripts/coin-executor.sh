#!/usr/bin/env bash
# coin-executor → Gitea coin/coin-executor + job coin-executor в Jenkins.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

EXEC="${REPO_ROOT}/coin-executor"
[[ -d "${EXEC}" ]] || { echo "missing ${EXEC}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-executor.git"
echo "==> Gitea: coin/coin-executor"
gitea_create_repo "coin-executor"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete \
  --exclude '/coin-executor' \
  "${EXEC}/" "${WORK}/"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-executor (local stack)"
git push --force "${url}" main

chmod +x "${EXEC}/scripts/"*.sh "${EXEC}/scripts/lib/"*.sh 2>/dev/null || true

echo "==> Jenkins: job coin-executor"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-executor-build.yaml" \
  /var/jenkins_home/casc-config/45-coin-executor-build.yaml

echo "coin-executor ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-executor"
