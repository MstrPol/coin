#!/usr/bin/env bash
# Пушит demo-python-uv (из coin-starters) в Gitea.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
STARTER="${REPO_ROOT}/coin-starters/python-uv-app"

# shellcheck disable=SC1091
[[ -f "${ROOT}/.env" ]] && source "${ROOT}/.env"

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
GITEA_URL="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/demo-python-uv.git"

if [[ ! -d "${STARTER}" ]]; then
  echo "starter not found: ${STARTER}" >&2
  exit 1
fi

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

rsync -a "${STARTER}/" "${WORK}/"

cat > "${WORK}/Jenkinsfile" <<'EOF'
@Library('coin-lib') _

coinPipeline(cloud: 'kubernetes')
EOF

cat > "${WORK}/.coin/config.yaml" <<'EOF'
version: 1

coin:
  template: python-uv-app
  templateVersion: v1

jenkins:
  credentials:
    docker: nexus-docker

project:
  name: demo-python-uv
  groupId: com.example.coin
  repository: Nexus_PROD

container:
  port: 8080
  command: ["python", "-m", "my_service"]

rn:
  serviceUrl: http://nexus:8081
EOF

cd "${WORK}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add -A
git commit -m "demo service for local stack"

git push --force "${GITEA_URL}" main

echo "demo-python-uv pushed to Gitea"
