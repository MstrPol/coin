#!/usr/bin/env bash
# Maven2 URL helpers (maven-releases / maven-snapshots only — no raw repos).

maven_repo_for_version() {
  case "$1" in
    *-SNAPSHOT) echo "${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}" ;;
    *) echo "${NEXUS_MAVEN_RELEASES:-maven-releases}" ;;
  esac
}

# gp_content_url NAME VERSION
gp_content_url() {
  local name="$1" version="$2"
  local repo base
  repo="$(maven_repo_for_version "${version}")"
  base="${NEXUS_URL:-http://nexus:8081}/repository/${repo}"
  echo "${base}/coin/gp-content/${name}/${version}/${name}-${version}.zip"
}
