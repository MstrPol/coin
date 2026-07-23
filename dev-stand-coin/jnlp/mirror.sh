#!/usr/bin/env bash
# Mirror jenkins/inbound-agent → локальный Nexus (для container jnlp).
set -euo pipefail

stand_base="$(cd "$(dirname "$0")/../../dev-stand-base" && pwd)"

# shellcheck disable=SC1091
set -a && source "${stand_base}/.env" && set +a

PLATFORM="${DOCKER_PLATFORM:-linux/arm64}"
NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"
JNLP_TAG="${JNLP_TAG:-3327.v868139a_d00e0-8}"
SOURCE_IMAGE="${JNLP_SOURCE_IMAGE:-jenkins/inbound-agent:${JNLP_TAG}}"

registry_host="localhost:${NEXUS_DOCKER_PORT}"
# внутри k8s/compose DNS — nexus:8082
image_local="${registry_host}/${NEXUS_DOCKER_REPO}/inbound-agent:${JNLP_TAG}"

echo "==> pull ${SOURCE_IMAGE} (${PLATFORM})"
docker pull --platform "${PLATFORM}" "${SOURCE_IMAGE}"

echo "==> tag ${image_local}"
docker tag "${SOURCE_IMAGE}" "${image_local}"

echo "==> login ${registry_host}"
echo "${NEXUS_DOCKER_PASSWORD}" | docker login "${registry_host}" \
  -u "${NEXUS_DOCKER_USER}" --password-stdin

echo "==> push ${image_local}"
docker push "${image_local}"

echo "OK: ${image_local}"
echo "    в pod: nexus:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}/inbound-agent:${JNLP_TAG}"
