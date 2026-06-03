#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

ENV_FILE="${ROOT}/.env"
if [[ ! -f "${ENV_FILE}" ]]; then
  cp "${ROOT}/.env.example" "${ENV_FILE}"
  echo "created ${ENV_FILE} from .env.example"
fi

# shellcheck disable=SC1091
source "${ENV_FILE}"

chmod +x "${ROOT}/scripts/"*.sh 2>/dev/null || true
chmod +x "${ROOT}/scripts/"*.sh

REGISTRY_PORT="${REGISTRY_PORT:-5050}"
export REGISTRY="localhost:${REGISTRY_PORT}"

if [[ -z "${DOCKER_PLATFORM:-}" ]]; then
  DOCKER_PLATFORM="$("${ROOT}/scripts/detect-platform.sh" --platform)"
  export DOCKER_PLATFORM
  echo "DOCKER_PLATFORM=${DOCKER_PLATFORM} (auto)"
fi

port_busy() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    lsof -iTCP:"${port}" -sTCP:LISTEN -n -P >/dev/null 2>&1
    return
  fi
  (echo >/dev/tcp/127.0.0.1/"${port}") >/dev/null 2>&1
}

for port_var in REGISTRY_PORT JENKINS_HTTP_PORT GITEA_HTTP_PORT NEXUS_HTTP_PORT K3S_DASHBOARD_PORT; do
  port="${!port_var}"
  if port_busy "${port}"; then
    echo "ERROR: порт ${port} (${port_var}) уже занят." >&2
    if [[ "${port_var}" == "REGISTRY_PORT" && "${port}" == "5000" ]]; then
      echo "На macOS 5000 часто занят AirPlay. Задайте REGISTRY_PORT=5050 в docker/.env" >&2
    fi
    exit 1
  fi
done

echo "==> [1/8] building Jenkins image"
docker compose build jenkins

echo "==> [2/8] starting registry, nexus, k3s, gitea"
docker compose up -d registry nexus k3s gitea

echo "==> waiting for registry"
for i in $(seq 1 30); do
  if curl -sf "http://localhost:${REGISTRY_PORT}/v2/" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> [3/8] k3s kubeconfig + Dashboard + Jenkins auth"
for i in $(seq 1 60); do
  if docker compose exec -T k3s test -f /output/kubeconfig.yaml; then
    break
  fi
  sleep 2
done

docker compose exec -T k3s sh -c '
  cp /output/kubeconfig.yaml /output/config
  sed -i "s|127.0.0.1|k3s|g" /output/config
  sed -i "s|localhost|k3s|g" /output/config
'

"${ROOT}/scripts/setup-dashboard.sh"
"${ROOT}/scripts/setup-jenkins-k8s-auth.sh"

echo "==> [4/8] Gitea + Nexus"
"${ROOT}/scripts/prepare-gitea.sh"
"${ROOT}/scripts/prepare-nexus.sh"

echo "==> [5/8] push coin monorepo + demo service to Gitea"
"${ROOT}/scripts/push-coin.sh"
"${ROOT}/scripts/prepare-demo.sh"

echo "==> [6/8] starting Jenkins"
docker compose up -d jenkins

echo "==> waiting for Jenkins"
JENKINS_PORT="${JENKINS_HTTP_PORT:-8080}"
for i in $(seq 1 90); do
  if curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER:-admin}:${JENKINS_ADMIN_PASSWORD:-admin}" \
    "http://localhost:${JENKINS_PORT}/login"; then
    break
  fi
  sleep 3
done

"${ROOT}/scripts/register-jenkins-k8s-endpoints.sh"

echo "==> [7/8] platform build (coin-cli → Nexus, agents → registry)"
"${ROOT}/scripts/trigger-platform-build.sh" || {
  echo "WARN: platform build не завершился — запустите вручную: make platform-build" >&2
}

cat <<EOF

Coin local stack is up.

  Jenkins:   http://localhost:${JENKINS_HTTP_PORT:-8080}  (${JENKINS_ADMIN_USER:-admin} / ${JENKINS_ADMIN_PASSWORD:-admin})
  Gitea:     http://localhost:${GITEA_HTTP_PORT:-3000}  (${GITEA_USER:-coin} / ${GITEA_PASSWORD:-coin})
  Registry:  http://localhost:${REGISTRY_PORT}
  Nexus:     http://localhost:${NEXUS_HTTP_PORT:-8081}  (admin / ${NEXUS_ADMIN_PASSWORD:-coin12345})
  k3s UI:    https://localhost:${K3S_DASHBOARD_PORT:-8443}  (Token — make dashboard-token)

Git repos:
  http://localhost:${GITEA_HTTP_PORT:-3000}/coin/coin
  http://localhost:${GITEA_HTTP_PORT:-3000}/coin/demo-python-uv

Platform jobs (Jenkins):
  coin-cli → Nexus coin-cli/dev/coin_linux_<arch>
  coin-agents (BUILD_ALL) -> registry:5000/coin/ci-*

Service smoke: Jenkins → coin-demo-python-uv → Build Now

Teardown: make down

EOF
