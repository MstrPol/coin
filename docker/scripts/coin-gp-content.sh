#!/usr/bin/env bash
# coin-gp-content → Gitea coin/coin-gp-content + job coin-gp-content в Jenkins.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

CONTENT="${REPO_ROOT}/coin-gp-content"
[[ -d "${CONTENT}" ]] || { echo "missing ${CONTENT}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-gp-content.git"

echo "==> Gitea: coin/coin-gp-content"
gitea_create_repo "coin-gp-content"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete \
  --exclude 'dist/' \
  "${CONTENT}/" "${WORK}/"

chmod +x "${WORK}/scripts/"*.sh "${WORK}/scripts/lib/"*.sh 2>/dev/null || true

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-gp-content (local stack)"
git push --force "${url}" main

echo "==> Jenkins: job coin-gp-content"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-gp-content-build.yaml" \
  /var/jenkins_home/casc-config/46-coin-gp-content-build.yaml

echo "coin-gp-content ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-gp-content"
