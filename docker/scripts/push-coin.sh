#!/usr/bin/env bash
# Пушит monorepo coin в Gitea (platform repo).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
GITEA_URL="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin.git"

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

rsync -a --delete \
  --exclude '.git' \
  --exclude 'docker/' \
  --exclude 'coin-cli/coin' \
  --exclude '.cursor' \
  "${REPO_ROOT}/" "${WORK}/"

cp "${ROOT}/images-local.yaml" "${WORK}/coin-lib/resources/images.yaml"
cp "${ROOT}/jenkins/PodTemplate.local.groovy" "${WORK}/coin-lib/src/org/coin/ci/PodTemplate.groovy"

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git commit -m "coin platform (local stack)"

git push --force "${GITEA_URL}" main

echo "coin monorepo pushed to Gitea"
