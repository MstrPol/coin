#!/usr/bin/env bash
set -euo pipefail

echo "==> coin standard test (java-maven)"
mvn -B test
