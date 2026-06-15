#!/usr/bin/env bash
# Deprecated: версия берётся из coin-api next-version (Jenkins BUMP).
# Локальный helper при отсутствии API.
# Usage: compute-version.sh [VERSION_FILE] [BUMP] [SNAPSHOT]
set -euo pipefail

VERSION_FILE="${1:-VERSION}"
BUMP="${2:-none}"
SNAPSHOT="${3:-true}"

if [[ ! -f "${VERSION_FILE}" ]]; then
  echo "missing ${VERSION_FILE}" >&2
  exit 1
fi

base="$(tr -d '[:space:]' < "${VERSION_FILE}")"
if [[ ! "${base}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "invalid semver in ${VERSION_FILE}: ${base}" >&2
  exit 1
fi

IFS=. read -r major minor patch <<< "${base}"
major="${major:-0}"
minor="${minor:-0}"
patch="${patch:-0}"

case "${BUMP}" in
  major) base="$((major + 1)).0.0" ;;
  minor) base="${major}.$((minor + 1)).0" ;;
  patch) base="${major}.${minor}.$((patch + 1))" ;;
  none) ;;
  *)
    echo "unknown bump: ${BUMP}" >&2
    exit 1
    ;;
esac

if [[ "${SNAPSHOT}" == "true" ]]; then
  echo "${base}-SNAPSHOT"
else
  echo "${base}"
fi
