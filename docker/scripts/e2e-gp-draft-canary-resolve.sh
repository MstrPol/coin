#!/usr/bin/env bash
# GP draft on canary line with draft gp-content pin → canary resolve succeeds.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_API_KEY:-dev-local-admin-key}"
ACTOR="${COIN_PUBLISH_ACTOR:-e2e-gp-draft-canary}"
GP="${COIN_GP_E2E_NAME:-go-app}"
GP_STABLE="${COIN_GP_STABLE_VERSION:-1.0.0}"
GP_DRAFT="${COIN_GP_DRAFT_VERSION:-1.0.2-draft}"
GC_NAME="${COIN_GP_CONTENT_NAME:-go-app}"
GC_DRAFT="${COIN_GC_DRAFT_VERSION:-9.9.9-gc-draft-e2e}"
AUTH=(-H "X-API-Key: ${KEY}" -H "Content-Type: application/json")

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
for cmd in curl jq; do need "${cmd}"; done

api_post() {
  local path="$1" body="$2"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X POST "${API}${path}" "${AUTH[@]}" -d "${body}")"
  if [[ "${code}" != "201" && "${code}" != "200" ]]; then
    echo "POST ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  cat "${tmp}"
  rm -f "${tmp}"
}

echo "==> coin-api ready"
curl -fsS "${API}/ready" >/dev/null

echo "==> ensure stable GP ${GP}@${GP_STABLE}"
curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_STABLE}" "${AUTH[@]}" >/dev/null

comp="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/versions/${GP_STABLE}" "${AUTH[@]}" | jq -c '.composition')"
agent="$(echo "${comp}" | jq -r '.[] | select(.type=="agent") | .version')"
bm="$(echo "${comp}" | jq -r '.[] | select(.type=="branching-model") | .version')"

echo "==> gp-content draft ${GC_NAME}@${GC_DRAFT}"
api_post "/v1/admin/components" "$(jq -n --arg n "${GC_NAME}" --arg a "${ACTOR}" \
  '{type: "gp-content", name: $n, actor: $a}')" >/dev/null 2>&1 || true
api_post "/v1/admin/components/gp-content/${GC_NAME}/versions/drafts" \
  "$(jq -n --arg v "${GC_DRAFT}" --arg a "${ACTOR}" '{version: $v, actor: $a}')" >/dev/null 2>&1 || true

gc_status="$(curl -fsS "${API}/v1/admin/components/gp-content/${GC_NAME}/versions/${GC_DRAFT}" "${AUTH[@]}" | jq -r '.status')"
if [[ "${gc_status}" != "draft" && "${gc_status}" != "published" ]]; then
  echo "FAIL: gp-content draft missing (${gc_status})" >&2
  exit 1
fi

draft_body="$(jq -n --arg ver "${GP_DRAFT}" --arg agent "${agent}" --arg gc "${GC_DRAFT}" --arg bm "${bm}" --arg a "${ACTOR}" \
  '{version: $ver, agentStackName: "coin-agent", gpContentName: $GC_NAME, branchingModelName: "trunk-based", composition: {agent: $agent, "gp-content": $gc, "branching-model": $bm}, actor: $a}')"

echo "==> GP draft ${GP}@${GP_DRAFT} with draft gp-content pin"
code="$(curl -sS -o /tmp/coin-e2e-gp-draft.json -w '%{http_code}' -X POST \
  "${API}/v1/admin/golden-paths/${GP}/drafts" "${AUTH[@]}" -d "${draft_body}")"
if [[ "${code}" != "201" && "${code}" != "409" ]]; then
  echo "FAIL: GP draft HTTP ${code}: $(cat /tmp/coin-e2e-gp-draft.json)" >&2
  exit 1
fi

curl -fsS -X PATCH "${API}/v1/admin/golden-paths/${GP}/catalog" "${AUTH[@]}" \
  -d "$(jq -n --arg stable "${GP_STABLE}" --arg canary "${GP_DRAFT}" --arg a "${ACTOR}" \
    '{latest: $stable, latestCanary: $canary, minimum: $stable, actor: $a}')" >/dev/null

echo "==> canary resolve (forceChannel=canary)"
preview="$(curl -fsS "${API}/v1/admin/golden-paths/${GP}/resolve-preview?pin=*&forceChannel=canary" \
  -H "X-API-Key: ${KEY}")"
channel="$(echo "${preview}" | jq -r '.channel')"
resolved="$(echo "${preview}" | jq -r '.resolvedVersion')"

if [[ "${channel}" != "canary" ]]; then
  echo "FAIL: expected canary channel, got ${channel}" >&2
  exit 1
fi
if [[ "${resolved}" != "${GP_DRAFT}" && "${resolved}" != "${GP_DRAFT%-draft}" ]]; then
  echo "FAIL: expected GP draft on canary line, got ${resolved}" >&2
  exit 1
fi

echo "OK: GP draft ${GP}@${GP_DRAFT} on canary line with draft gp-content pin resolves (channel=${channel})"
