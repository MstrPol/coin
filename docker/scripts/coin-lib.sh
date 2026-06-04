#!/usr/bin/env bash
# coin-lib → Gitea coin/coin-lib + global Pipeline Library в Jenkins.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

COIN_LIB="${REPO_ROOT}/coin-lib"
[[ -d "${COIN_LIB}" ]] || { echo "missing ${COIN_LIB}" >&2; exit 1; }

echo "==> Gitea: coin/coin-lib"
gitea_create_repo "coin-lib"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete "${COIN_LIB}/" "${WORK}/"
cp "${ROOT}/PodTemplate.local.groovy" "${WORK}/src/org/coin/ci/PodTemplate.groovy"

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-lib.git"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-lib (local stack)"
git push --force "${url}" main

echo "==> Jenkins: global library coin-lib"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-lib.yaml" \
  /var/jenkins_home/casc-config/20-coin-lib.yaml

echo "coin-lib ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-lib"
