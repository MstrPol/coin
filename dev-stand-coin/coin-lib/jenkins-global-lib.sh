#!/usr/bin/env bash
set -euo pipefail

stand_coin="$(cd "$(dirname "$0")/.." && pwd)"
stand_base="$(cd "${stand_coin}/../dev-stand-base" && pwd)"
casc="${stand_coin}/jenkins/casc-coin-lib.yaml"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

for _ in $(seq 1 60); do
  curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" \
    "http://localhost:${JENKINS_HTTP_PORT}/login" && break
  sleep 3
done
curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER}:${JENKINS_ADMIN_PASSWORD}" \
  "http://localhost:${JENKINS_HTTP_PORT}/login" || {
  echo "Jenkins не отвечает на http://localhost:${JENKINS_HTTP_PORT}" >&2
  exit 1
}

jenkins_container="$(
  docker ps --filter "publish=${JENKINS_HTTP_PORT}" --format '{{.Names}}' | head -1
)"
[[ -n "${jenkins_container}" ]] || {
  echo "Jenkins-контейнер не найден (порт ${JENKINS_HTTP_PORT})" >&2
  exit 1
}

docker exec -i -u 0 "${jenkins_container}" sh -c \
  "cat > /var/jenkins_home/casc-config/47-coin-lib.yaml && chown jenkins:jenkins /var/jenkins_home/casc-config/47-coin-lib.yaml" \
  < "${casc}"

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

docker exec -T "${jenkins_container}" rm -rf /var/jenkins_home/caches/git-* 2>/dev/null || true

echo "Jenkins Global Library: coin-lib"
