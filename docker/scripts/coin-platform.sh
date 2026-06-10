#!/usr/bin/env bash
# PF-16: push split platform repos to local Gitea (agents + starters).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPTS="${ROOT}/scripts"

"${SCRIPTS}/coin-jenkins-agents.sh"
"${SCRIPTS}/coin-starters.sh"

echo "==> platform repos push complete (coin-jenkins-agents + coin-starters)"
