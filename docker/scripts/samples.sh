#!/usr/bin/env bash
# Примеры продуктовых репо: starters → samples/ (gitignored) → Gitea + multibranch jobs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
MANIFEST="${ROOT}/samples.yaml"
SAMPLES_DIR="${REPO_ROOT}/samples"
STARTERS="${REPO_ROOT}/coin-starters"

# shellcheck source=lib/common.sh
source "${LIB}"
load_env

# samples/ — отдельные git-репо (Gitea). Monorepo (GitHub) этим скриптом не коммитится.
if [[ -n "${GIT_DIR:-}" || -n "${GIT_WORK_TREE:-}" ]]; then
  echo "GIT_DIR/GIT_WORK_TREE заданы — отмена (риск коммита в monorepo вместо samples/)" >&2
  exit 1
fi
if git -C "${REPO_ROOT}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  mono_origin="$(git -C "${REPO_ROOT}" config --get remote.origin.url 2>/dev/null || true)"
  if [[ "${mono_origin}" == *"localhost:"* ]] || [[ "${mono_origin}" == *"/coin/demo-"* ]]; then
    echo "origin монорепо указывает на Gitea demo — задайте GitHub (git@github.com:MstrPol/coin.git)" >&2
    exit 1
  fi
fi

[[ -f "${MANIFEST}" ]] || { echo "missing ${MANIFEST}" >&2; exit 1; }

GITEA_HTTP_PORT="${GITEA_HTTP_PORT:-3000}"
GITEA_USER="${GITEA_USER:-coin}"
GITEA_PASSWORD="${GITEA_PASSWORD:-coin}"

mkdir -p "${SAMPLES_DIR}"

REPOS=()

write_jenkinsfile() {
  local dest="$1"
  cp "${STARTERS}/Jenkinsfile.coin" "${dest}/Jenkinsfile"
}

patch_sample_project() {
  local dest="$1"
  local project="$2"
  local cfg="${dest}/.coin/config.yaml"

  sed -i.bak "s/^  name: .*/  name: ${project}/" "${cfg}"
  sed -i.bak "s/^  artifactId: .*/  artifactId: ${project}/" "${cfg}"
  sed -i.bak 's|^const serviceName = .*|const serviceName = "'"${project}"'"|' "${dest}/main.go" 2>/dev/null || true
  if [[ -f "${dest}/go.mod" ]]; then
    sed -i.bak "s|^module .*|module example.com/${project}|" "${dest}/go.mod"
    rm -f "${dest}/go.mod.bak"
  fi
  rm -f "${cfg}.bak" "${dest}/main.go.bak"
}

ensure_sample_git_repo() {
  local dest="$1"
  if [[ -d "${dest}/.git" ]] && ! git -C "${dest}" rev-parse --git-dir >/dev/null 2>&1; then
    rm -rf "${dest}/.git"
  fi
  if ! git -C "${dest}" rev-parse --is-inside-work-tree >/dev/null 2>&1 || \
     [[ "$(git -C "${dest}" rev-parse --show-toplevel)" != "${dest}" ]]; then
    rm -rf "${dest}/.git"
    git -C "${dest}" init -b main
    git -C "${dest}" config user.email "coin@local"
    git -C "${dest}" config user.name "Coin Local"
  fi
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
                traits {
                  gitBranchDiscovery()
                  gitTagDiscovery()
                }
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
    patch_sample_project "${dest}" "${repo}"

    url="http://${GITEA_USER}:${GITEA_PASSWORD}@localhost:${GITEA_HTTP_PORT}/coin/${repo}.git"

    ensure_sample_git_repo "${dest}"
    if git -C "${dest}" remote get-url origin >/dev/null 2>&1; then
      git -C "${dest}" remote set-url origin "${url}"
    else
      git -C "${dest}" remote add origin "${url}"
    fi
    git -C "${dest}" add -A
    git -C "${dest}" diff --staged --quiet || git -C "${dest}" commit -m "sample product repo (${starter})"
    git -C "${dest}" push --force -u origin main

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

echo "samples ready under ${SAMPLES_DIR}/"
for repo in "${REPOS[@]}"; do
  echo "  job: ${repo} → http://gitea:3000/coin/${repo}.git"
done
