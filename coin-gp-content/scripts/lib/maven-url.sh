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

# gp_content_artifact_url NAME VERSION ARTIFACT_KEY
# Must match coin-api nexus.ArtifactMavenCoords.
gp_content_artifact_url() {
  local name="$1" version="$2" key="$3"
  local repo base coords classifier ext file
  repo="$(maven_repo_for_version "${version}")"
  base="${NEXUS_URL:-http://nexus:8081}/repository/${repo}"
  coords="$(python3 -c '
import sys
key = sys.argv[1]
known = {"json", "sh", "yaml", "yml", "md", "txt"}
normalized = key.replace("/", ".").replace(" ", "_")
if "." in key:
    candidate = key.rsplit(".", 1)[-1].lower()
    if candidate in known:
        dot = normalized.rfind(".")
        print(f"{normalized[:dot]} {candidate}")
        raise SystemExit
base = key.rsplit("/", 1)[-1]
if base.lower() == "containerfile":
    print(f"{normalized} containerfile")
else:
    print(f"{normalized} ")
' "${key}")"
  classifier="${coords%% *}"
  ext="${coords#* }"
  file="${name}-${version}"
  if [[ -n "${classifier}" ]]; then
    file+="-${classifier}"
  fi
  if [[ -n "${ext}" ]]; then
    file+=".${ext}"
  fi
  echo "${base}/coin/gp/content/${name}/${version}/${file}"
}
