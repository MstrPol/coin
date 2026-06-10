#!/usr/bin/env bash
# Подъём инфраструктуры: nexus (+ docker repo), k3s, gitea, jenkins (plugins + k8s agents).
# Platform jobs — make coin-platform / coin-executor / samples. k3s Dashboard — make dashboard.
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

chmod +x "${ROOT}/scripts/"*.sh "${ROOT}/scripts/lib/"*.sh 2>/dev/null || true

NEXUS_DOCKER_PORT="${NEXUS_DOCKER_PORT:-8082}"

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

for port_var in NEXUS_DOCKER_PORT JENKINS_HTTP_PORT GITEA_HTTP_PORT NEXUS_HTTP_PORT K3S_DASHBOARD_PORT; do
  port="${!port_var}"
  if port_busy "${port}"; then
    echo "ERROR: порт ${port} (${port_var}) уже занят." >&2
    exit 1
  fi
done

build_jenkins_if_needed() {
  local image="${COMPOSE_PROJECT_NAME:-coin}-jenkins"
  if [[ "${FORCE_JENKINS_BUILD:-0}" == "1" ]]; then
    echo "==> [1/6] building Jenkins image (FORCE_JENKINS_BUILD=1)..."
  elif docker image inspect "${image}" >/dev/null 2>&1; then
    echo "==> [1/6] Jenkins image ${image} — skip build"
    echo "    пересборка: docker compose build jenkins"
    return 0
  else
    echo "==> [1/6] building Jenkins image (первый раз ~5–15 мин — загрузка plugins, будет verbose-лог)..."
  fi
  DOCKER_BUILDKIT=1 docker compose --progress plain build jenkins
}

build_jenkins_if_needed

echo "==> [2/6] starting k3s"
docker compose up -d k3s

for _ in $(seq 1 60); do
  if docker compose exec -T k3s test -f /output/kubeconfig.yaml 2>/dev/null; then
    break
  fi
  sleep 2
done

echo "==> [2/6] starting nexus, gitea"
docker compose up -d nexus gitea

echo "==> [3/6] k3s auth (Jenkins SA token)"

docker compose exec -T k3s sh -c '
  cp /output/kubeconfig.yaml /output/config
  sed -i "s|127.0.0.1|k3s|g" /output/config
  sed -i "s|localhost|k3s|g" /output/config
'

"${ROOT}/scripts/setup-jenkins-k8s-auth.sh"

echo "==> [4/6] init Gitea + Nexus (admin, Docker repo, k8s Endpoints)"
"${ROOT}/scripts/prepare-gitea.sh"
"${ROOT}/scripts/prepare-nexus.sh"

echo "==> [5/6] starting Jenkins (plugins + Kubernetes cloud для dynamic agents)"
docker compose up -d jenkins

echo "==> [6/6] waiting for Jenkins + finalize k3s"
JENKINS_PORT="${JENKINS_HTTP_PORT:-8080}"
for _ in $(seq 1 90); do
  if curl -sf -o /dev/null -u "${JENKINS_ADMIN_USER:-admin}:${JENKINS_ADMIN_PASSWORD:-admin}" \
    "http://localhost:${JENKINS_PORT}/login"; then
    break
  fi
  sleep 3
done

"${ROOT}/scripts/fix-k3s.sh"

cat <<EOF

Coin local stack is up (infrastructure only).

  Jenkins:   http://localhost:${JENKINS_HTTP_PORT:-8080}  (${JENKINS_ADMIN_USER:-admin} / ${JENKINS_ADMIN_PASSWORD:-admin})
             plugins + Kubernetes cloud + creds (k3s, gitea, nexus)
  Gitea:     http://localhost:${GITEA_HTTP_PORT:-3000}  (${GITEA_USER:-coin} / ${GITEA_PASSWORD:-coin})
  Nexus:     http://localhost:${NEXUS_HTTP_PORT:-8081}  (admin / ${NEXUS_ADMIN_PASSWORD:-coin12345})
  Docker:    localhost:${NEXUS_DOCKER_PORT}/${NEXUS_DOCKER_REPO:-coin-docker}  (${NEXUS_DOCKER_USER:-coin} / ${NEXUS_DOCKER_PASSWORD:-coin})
  k3s UI:    make dashboard  (опционально)

Platform (после bootstrap):
  make coin-jenkins-agents   # agents → Gitea coin/coin-jenkins-agents
  make coin-starters         # starters → Gitea coin/coin-starters
  make coin-platform         # both (PF-16 meta)
  make coin-executor       # coin-executor → Gitea + job coin-executor
  make agents-build        # job agents-build (JCasC)
  make samples             # demo-продукты → samples/ + Gitea

Teardown: make down  |  make reset

EOF
