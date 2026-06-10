#!/usr/bin/env bash
# Jenkins job agents-build (coin-jenkins-agents/Jenkinsfile).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

echo "==> Jenkins: job agents-build"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-agents-build.yaml" \
  /var/jenkins_home/casc-config/30-agents-build.yaml

echo "agents-build job ready"
