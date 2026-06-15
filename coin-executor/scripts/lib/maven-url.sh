#!/usr/bin/env bash
# Maven2 URL helpers for coin-executor binary.

executor_binary_url() {
  local version="$1" goarch="$2"
  local repo base
  if echo "${version}" | grep -q SNAPSHOT; then
    repo="${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}"
  else
    repo="${NEXUS_MAVEN_RELEASES:-maven-releases}"
  fi
  base="${NEXUS_URL:-http://nexus:8081}/repository/${repo}"
  echo "${base}/coin/executor/coin-executor/${version}/coin-executor-${version}-linux-${goarch}.bin"
}
