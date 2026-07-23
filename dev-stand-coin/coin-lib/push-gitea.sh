#!/usr/bin/env bash
set -euo pipefail

stand_base="$(cd "$(dirname "$0")/../../dev-stand-base" && pwd)"
coin_lib="$(cd "$(dirname "$0")/../../../coin-lib" && pwd)"
remote="gitea-local"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

[[ -d "${coin_lib}/.git" ]] || { echo "нет репозитория: ${coin_lib}" >&2; exit 1; }

url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-lib.git"

curl -sf -u "${GITEA_USER}:${GITEA_PASSWORD}" \
  -X POST "http://localhost:${GITEA_HTTP_PORT}/api/v1/user/repos" \
  -H "Content-Type: application/json" \
  -d '{"name":"coin-lib","private":false,"default_branch":"main"}' \
  >/dev/null 2>&1 || true

cd "${coin_lib}"
if git remote get-url "${remote}" >/dev/null 2>&1; then
  git remote set-url "${remote}" "${url}"
else
  git remote add "${remote}" "${url}"
fi

git push --force "${remote}" --all
git push --force "${remote}" --tags
git remote remove "${remote}"

echo "Gitea: http://localhost:${GITEA_HTTP_PORT}/coin/coin-lib"
