#!/usr/bin/env bash
# DEPRECATED: Gitea tag retriever for coin-lib — bootstrap only.
# Primary path: publish-lib.sh → Nexus ZIP + make coin-lib-http (или make seed-jenkins-lib).
# coin-lib → Gitea coin/coin-lib (tag 1.0.0) + Jenkins Global Shared Library.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

COIN_LIB="${REPO_ROOT}/coin-lib"
[[ -d "${COIN_LIB}" ]] || { echo "missing ${COIN_LIB}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
LIB_TAG="1.0.0"
url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-lib.git"

echo "==> Gitea: coin/coin-lib (tag ${LIB_TAG})"
gitea_create_repo "coin-lib"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
rsync -a --delete "${COIN_LIB}/" "${WORK}/"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git diff --staged --quiet || git commit -m "coin-lib (local stack)"
git tag -f "${LIB_TAG}"
git push --force "${url}" main
git push --force "${url}" "refs/tags/${LIB_TAG}"

echo "==> Jenkins: Global Shared Library coin-lib@${LIB_TAG}"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-lib.yaml" \
  /var/jenkins_home/casc-config/47-coin-lib.yaml
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-lib-build.yaml" \
  /var/jenkins_home/casc-config/48-coin-lib-build.yaml

# Force-push tag: Jenkins library cache may keep stale tag→commit mapping.
docker compose -f "${ROOT}/compose.yml" exec -T jenkins \
  rm -rf /var/jenkins_home/caches/git-* 2>/dev/null || true

echo "coin-lib ready: http://localhost:${GITEA_HTTP_PORT}/coin/coin-lib"
echo "  Jenkins Global Library: coin-lib@${LIB_TAG}"
echo "  Jenkins job: coin-lib (publish lib ZIP to Nexus)"
echo "  Phase 2 HTTP lib: make coin-lib-http"
