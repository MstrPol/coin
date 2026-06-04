#!/usr/bin/env bash
# coin-cli → Gitea coin/coin-cli + job coin-cli в Jenkins.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

CLI="${REPO_ROOT}/coin-cli"
[[ -d "${CLI}" ]] || { echo "missing ${CLI}" >&2; exit 1; }

echo "==> Gitea: coin/coin-cli"
gitea_create_repo "coin-cli"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete \
  --exclude 'coin' \
  --exclude '.coin/cache/' \
  --exclude '.coin/temp/' \
  --exclude '.coin/generated/' \
  "${CLI}/" "${WORK}/"

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-cli.git"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-cli (local stack)"
git push --force "${url}" main

echo "==> Jenkins: job coin-cli"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-cli-build.yaml" \
  /var/jenkins_home/casc-config/40-coin-cli-build.yaml

echo "coin-cli ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-cli"
