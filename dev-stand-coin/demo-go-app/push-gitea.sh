#!/usr/bin/env bash
# Скопировать coin/samples/demo-go-app → Gitea coin/demo-go-app (force main).
set -euo pipefail

stand_coin="$(cd "$(dirname "$0")/.." && pwd)"
stand_base="$(cd "${stand_coin}/../dev-stand-base" && pwd)"
sample="$(cd "${stand_coin}/../samples/demo-go-app" && pwd)"
repo_name="${DEMO_GO_APP_REPO:-demo-go-app}"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"

[[ -d "${sample}" ]] || {
  echo "нет sample: ${sample}" >&2
  exit 1
}

echo "==> Gitea: ${GITEA_USER}/${repo_name}"
curl -sf -u "${GITEA_USER}:${GITEA_PASSWORD}" \
  -X POST "http://localhost:${GITEA_HTTP_PORT}/api/v1/user/repos" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${repo_name}\",\"private\":false,\"default_branch\":\"main\"}" \
  >/dev/null 2>&1 || true

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

# Содержимое working tree без .git sample-репо
rsync -a --exclude '.git' "${sample}/" "${work}/"

cd "${work}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git commit -m "demo-go-app: sync from coin/samples/demo-go-app"

push_url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/${GITEA_USER}/${repo_name}.git"
git push --force "${push_url}" main

echo "OK: http://localhost:${GITEA_HTTP_PORT}/${GITEA_USER}/${repo_name}"
