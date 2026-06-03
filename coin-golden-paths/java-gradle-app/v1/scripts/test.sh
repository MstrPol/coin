#!/usr/bin/env bash
set -euo pipefail

echo "==> coin standard test (java-gradle)"
if [[ -x ./gradlew ]]; then
  ./gradlew test --no-daemon
else
  gradle test --no-daemon
fi
