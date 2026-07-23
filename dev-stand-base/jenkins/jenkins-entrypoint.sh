#!/usr/bin/env bash
set -euo pipefail

KUBECONFIG_PATH="${KUBECONFIG_PATH:-/kubeconfig/config}"
JENKINS_TOKEN_PATH="${JENKINS_TOKEN_PATH:-/kubeconfig/jenkins-token}"
CASC_DIR="/var/jenkins_home/casc-config"

resolve_kubeconfig() {
  if [[ -f "${KUBECONFIG_PATH}" ]]; then
    echo "${KUBECONFIG_PATH}"
    return
  fi
  if [[ -f /kubeconfig/kubeconfig.yaml ]]; then
    echo /kubeconfig/kubeconfig.yaml
    return
  fi
  echo ""
}

for i in $(seq 1 120); do
  if [[ -n "$(resolve_kubeconfig)" && -f "${JENKINS_TOKEN_PATH}" ]]; then
    break
  fi
  echo "waiting for kubeconfig and jenkins token in /kubeconfig..."
  sleep 2
done

if [[ -z "$(resolve_kubeconfig)" ]]; then
  echo "kubeconfig not found in /kubeconfig (expected config or kubeconfig.yaml)" >&2
  exit 1
fi
if [[ ! -f "${JENKINS_TOKEN_PATH}" ]]; then
  echo "jenkins token not found: ${JENKINS_TOKEN_PATH} (run: make bootstrap)" >&2
  exit 1
fi

export K3S_TOKEN="$(tr -d '\n\r' < "${JENKINS_TOKEN_PATH}")"

mkdir -p "${CASC_DIR}"
cp /usr/share/jenkins/ref/casc.yaml "${CASC_DIR}/00-base.yaml"

# docker.sock с хоста: gid часто ≠ группа docker в контейнере (Debian/Mac).
if [[ -S /var/run/docker.sock ]]; then
  docker_gid="$(stat -c '%g' /var/run/docker.sock)"
  if ! getent group "${docker_gid}" >/dev/null; then
    groupadd -g "${docker_gid}" dockersock
  fi
  usermod -aG "${docker_gid}" jenkins
fi

export CASC_JENKINS_CONFIG="${CASC_DIR}"
exec setpriv --reuid=jenkins --regid=jenkins --init-groups /usr/local/bin/jenkins.sh
