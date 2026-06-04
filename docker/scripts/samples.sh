#!/usr/bin/env bash
# Примеры продуктовых репо: starters → samples/ (gitignored) → Gitea + multibranch jobs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
MANIFEST="${ROOT}/samples.yaml"
SAMPLES_DIR="${REPO_ROOT}/samples"
STARTERS="${REPO_ROOT}/coin-platform/starters"

# shellcheck source=lib/common.sh
source "${LIB}"
load_env

[[ -f "${MANIFEST}" ]] || { echo "missing ${MANIFEST}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"

mkdir -p "${SAMPLES_DIR}"

REPOS=()

write_jenkinsfile() {
  local dest="$1"
  cat > "${dest}/Jenkinsfile" <<'EOF'
@Library('coin-lib') _

coinPipeline(cloud: 'kubernetes')
EOF
}

patch_config() {
  local cfg="$1"
  local project="$2"
  sed -i.bak "s/^  name: .*/  name: ${project}/" "${cfg}"
  sed -i.bak 's|^  serviceUrl: .*|  serviceUrl: http://nexus:8081|' "${cfg}"
  rm -f "${cfg}.bak"
}

render_multibranch_casc() {
  local out="$1"
  cat > "${out}" <<'EOF'
jobs:
EOF
  local repo
  for repo in "${REPOS[@]}"; do
    cat >> "${out}" <<EOF
  - script: >
      multibranchPipelineJob('${repo}') {
        description('Sample product: ${repo}')
        branchSources {
          branchSource {
            source {
              git {
                id('${repo}')
                remote('http://gitea:3000/coin/${repo}.git')
                credentialsId('gitea-git')
              }
            }
          }
        }
        factory {
          workflowBranchProjectFactory {
            scriptPath('Jenkinsfile')
          }
        }
        triggers {
          periodicFolderTrigger {
            interval('15m')
          }
        }
        orphanedItemStrategy {
          defaultOrphanedItemStrategy {
            pruneDeadBranches(true)
            daysToKeepStr('-1')
            numToKeepStr('5')
          }
        }
      }
EOF
  done
}

repo=""
starter=""
while IFS= read -r line; do
  line="${line%%#*}"
  line="${line#"${line%%[![:space:]]*}"}"
  [[ -z "${line}" ]] && continue
  line="${line#- }"

  if [[ "${line}" =~ ^repo:[[:space:]]*(.+)$ ]]; then
    repo="${BASH_REMATCH[1]}"
    continue
  fi
  if [[ "${line}" =~ ^starter:[[:space:]]*(.+)$ ]]; then
    starter="${BASH_REMATCH[1]}"
    [[ -z "${repo}" || -z "${starter}" ]] && continue

    src="${STARTERS}/${starter}"
    dest="${SAMPLES_DIR}/${repo}"

    [[ -d "${src}" ]] || { echo "starter not found: ${src}" >&2; exit 1; }

    echo "==> sample ${repo} (from ${starter})"
    gitea_create_repo "${repo}"
    REPOS+=("${repo}")

    rsync -a --delete --exclude '.git' "${src}/" "${dest}/"
    write_jenkinsfile "${dest}"
    patch_config "${dest}/.coin/config.yaml" "${repo}"

    url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/${repo}.git"

    cd "${dest}"
    if [[ ! -d .git ]]; then
      git init -b main
      git config user.email "coin@local"
      git config user.name "Coin Local"
    fi
    if git remote get-url origin >/dev/null 2>&1; then
      git remote set-url origin "${url}"
    else
      git remote add origin "${url}"
    fi
    git add -A
    git diff --staged --quiet || git commit -m "sample product repo (${starter})"
    git push --force -u origin main

    echo "    http://localhost:${GITEA_HTTP_PORT}/coin/${repo}"

    repo=""
    starter=""
  fi
done < "${MANIFEST}"

if [[ ${#REPOS[@]} -eq 0 ]]; then
  echo "no samples pushed (check ${MANIFEST})" >&2
  exit 1
fi

echo "==> Jenkins: multibranch jobs (${#REPOS[@]})"
jenkins_wait
CASC="$(mktemp)"
trap 'rm -f "${CASC}"' RETURN
render_multibranch_casc "${CASC}"
jenkins_casc_reload "${CASC}" /var/jenkins_home/casc-config/35-samples-jobs.yaml

echo "==> Jenkins: branch scan"
for repo in "${REPOS[@]}"; do
  jenkins_build_job "${repo}" || true
done

echo "samples ready under ${SAMPLES_DIR}/"
for repo in "${REPOS[@]}"; do
  echo "  job: ${repo} → http://gitea:3000/coin/${repo}.git"
done
