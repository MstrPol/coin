#!/usr/bin/env bash
# Создать repo coin-smoke в Gitea + Jenkins job coin-smoke (dynamic agent smoke).
set -euo pipefail

stand_coin="$(cd "$(dirname "$0")/.." && pwd)"
stand_base="$(cd "${stand_coin}/../dev-stand-base" && pwd)"
smoke_dir="$(cd "$(dirname "$0")" && pwd)"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"
JENKINS_HTTP_PORT="${JENKINS_HTTP_PORT:-8080}"
JENKINS_ADMIN_USER="${JENKINS_ADMIN_USER:-admin}"
JENKINS_ADMIN_PASSWORD="${JENKINS_ADMIN_PASSWORD:-admin}"

push_url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/coin-smoke.git"

echo "==> Gitea: coin/coin-smoke"
curl -sf -u "${GITEA_USER}:${GITEA_PASSWORD}" \
  -X POST "http://localhost:${GITEA_HTTP_PORT}/api/v1/user/repos" \
  -H "Content-Type: application/json" \
  -d '{"name":"coin-smoke","private":false,"default_branch":"main"}' \
  >/dev/null 2>&1 || true

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
cp "${smoke_dir}/Jenkinsfile" "${work}/Jenkinsfile"
cd "${work}"
git init -b main
git config user.email "coin@local"
git config user.name "Coin Local"
git add Jenkinsfile
git commit -m "coin-smoke: dynamic agent smoke"
git push --force "${push_url}" main

echo "==> waiting for Jenkins"
for _ in $(seq 1 60); do
  curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" \
    "http://localhost:${JENKINS_HTTP_PORT}/login" && break
  sleep 3
done

jenkins_container="$(
  docker ps --filter "publish=${JENKINS_HTTP_PORT}" --format '{{.Names}}' | head -1
)"
[[ -n "${jenkins_container}" ]] || {
  echo "Jenkins-контейнер не найден" >&2
  exit 1
}

echo "==> Jenkins job coin-smoke (CASC)"
docker exec -i -u 0 "${jenkins_container}" sh -c \
  "cat > /var/jenkins_home/casc-config/49-coin-smoke.yaml && chown jenkins:jenkins /var/jenkins_home/casc-config/49-coin-smoke.yaml" \
  < "${smoke_dir}/casc-job.yaml"

cookie="$(mktemp)"
crumb="$(curl -sf -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" -c "${cookie}" \
  "http://localhost:${JENKINS_HTTP_PORT}/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)")"

http_code="$(curl -sf -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" -b "${cookie}" -c "${cookie}" \
  -H "${crumb}" -o /dev/null -w '%{http_code}' \
  -X POST "http://localhost:${JENKINS_HTTP_PORT}/configuration-as-code/reload")"
rm -f "${cookie}"

[[ "${http_code}" == "200" || "${http_code}" == "302" ]] || {
  echo "Jenkins CASC reload failed (HTTP ${http_code})" >&2
  exit 1
}

echo "OK: job http://localhost:${JENKINS_HTTP_PORT}/job/coin-smoke/"
echo "    перед запуском: make publish-jnlp && make publish-executor && make coin-lib"
echo "    Build with Parameters → COIN_AGENT_VERSION=1.0.0"
