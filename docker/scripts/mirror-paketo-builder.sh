#!/usr/bin/env bash
# One-time: mirror Paketo builder to Nexus for buildpack E2E (local registry, no Docker Hub in pod).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib/common.sh
source "${ROOT}/scripts/lib/common.sh"
load_env

NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"
NEXUS_DOCKER_REPO="${NEXUS_DOCKER_REPO:-coin-docker}"
NEXUS_DOCKER_USER="${NEXUS_DOCKER_USER:-coin}"
NEXUS_DOCKER_PASSWORD="${NEXUS_DOCKER_PASSWORD:-coin1234}"
UPSTREAM="${PAKETO_BUILDER:-paketobuildpacks/builder-jammy-base:latest}"
TARGET="localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO}/paketo-builder-jammy-base:latest"

if [[ -z "${DOCKER_PLATFORM:-}" ]]; then
  DOCKER_PLATFORM="linux/$(bash "${ROOT}/scripts/detect-platform.sh" --goarch)"
fi

echo "==> pull ${UPSTREAM} (${DOCKER_PLATFORM})"
docker pull --platform "${DOCKER_PLATFORM}" "${UPSTREAM}"
docker tag "${UPSTREAM}" "${TARGET}"
echo "${NEXUS_DOCKER_PASSWORD}" | docker login "localhost:${NEXUS_DOCKER_PORT}" \
  -u "${NEXUS_DOCKER_USER}" --password-stdin
echo "==> push ${TARGET}"
docker push "${TARGET}"
echo "OK: ${TARGET} platform=${DOCKER_PLATFORM}"
