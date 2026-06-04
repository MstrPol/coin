#!/usr/bin/env bash
# Примеры продуктовых репо: starters → samples/ (gitignored) → Gitea coin/<repo>.
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

    rsync -a --delete "${src}/" "${dest}/"
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

if [[ -z "$(ls -A "${SAMPLES_DIR}" 2>/dev/null)" ]]; then
  echo "no samples pushed (check ${MANIFEST})" >&2
  exit 1
fi

echo "samples ready under ${SAMPLES_DIR}/"
