#!/usr/bin/env bash
# coin-starters → Gitea coin/coin-starters
set -euo pipefail
SCRIPTS="$(cd "$(dirname "$0")" && pwd)"
"${SCRIPTS}/coin-starters.sh"
echo "==> platform repos push complete (coin-starters)"
