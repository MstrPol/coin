#!/usr/bin/env bash
# Переключить Jenkins Global Library coin-lib на Nexus HTTP retriever (target Phase 2).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

echo "==> Jenkins: coin-lib HTTP retriever (Nexus ZIP)"
jenkins_wait
jenkins_casc_reload \
  "${ROOT}/platform/jenkins/casc-coin-lib-http.yaml" \
  /var/jenkins_home/casc-config/47-coin-lib.yaml

echo "coin-lib HTTP retriever active (требует lib/coin-lib@version в Nexus)"
