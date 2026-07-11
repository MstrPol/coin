#!/usr/bin/env bash
# Maven2 URL helpers for branching-model component packages.

maven_repo_for_version() {
  case "$1" in
    *-SNAPSHOT) echo "${NEXUS_MAVEN_SNAPSHOTS:-maven-snapshots}" ;;
    *) echo "${NEXUS_MAVEN_RELEASES:-maven-releases}" ;;
  esac
}

# branching_model_package_manifest_url NAME VERSION
branching_model_package_manifest_url() {
  local name="$1" version="$2"
  local repo base
  repo="$(maven_repo_for_version "${version}")"
  base="${NEXUS_URL:-http://nexus:8081}/repository/${repo}"
  echo "${base}/coin/branching/model/${name}/${version}/${name}-${version}-package.manifest.json"
  # groupId coin.branching.model — см. coin-api nexus.ComponentPackageGroupID
}
