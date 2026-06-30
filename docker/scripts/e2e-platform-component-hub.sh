#!/usr/bin/env bash
# Platform component hub: agent draft lifecycle + branching-model profile draft.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LIB="${ROOT}/scripts/lib/common.sh"
# shellcheck source=lib/common.sh
source "${LIB}"
load_env

API="${COIN_API_URL:-http://localhost:8090}"
KEY="${COIN_API_KEY:-dev-local-admin-key}"
ACTOR="${COIN_PUBLISH_ACTOR:-e2e-platform-hub}"
AUTH=(-H "X-API-Key: ${KEY}" -H "Content-Type: application/json")
AGENT_VERSION="${COIN_AGENT_E2E_VERSION:-9.9.9-hub-e2e}"
BML_NAME="${COIN_BML_HUB_NAME:-hub-e2e-model}"
BML_VERSION="${COIN_BML_HUB_VERSION:-0.1.0-hub-draft}"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
for cmd in curl jq; do need "${cmd}"; done

echo "==> coin-api ready"
curl -fsS "${API}/ready" >/dev/null

api_post() {
  local path="$1" body="$2"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X POST "${API}${path}" "${AUTH[@]}" -d "${body}")"
  if [[ "${code}" != "201" && "${code}" != "200" && "${code}" != "409" ]]; then
    echo "POST ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  cat "${tmp}"
  rm -f "${tmp}"
}

api_delete() {
  local path="$1"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "${tmp}" -w '%{http_code}' -X DELETE "${API}${path}" "${AUTH[@]}")"
  if [[ "${code}" != "204" && "${code}" != "404" ]]; then
    echo "DELETE ${path} failed HTTP ${code}: $(cat "${tmp}")" >&2
    rm -f "${tmp}"
    exit 1
  fi
  rm -f "${tmp}"
  echo "${code}"
}

echo "==> agent draft register ${AGENT_VERSION}"
api_post "/v1/admin/components" \
  "$(jq -n --arg a "${ACTOR}" '{type:"agent",name:"coin-agent",actor:$a}')" >/dev/null
E2E_DIGEST="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
api_post "/v1/admin/components/agent/coin-agent/versions/drafts" \
  "$(jq -n --arg v "${AGENT_VERSION}" --arg a "${ACTOR}" --arg d "${E2E_DIGEST}" \
    '{version:$v, metadata:{image:"nexus:8082/coin-docker/coin-agent:'"${AGENT_VERSION}"'",digest:$d,runtime:"coin-agent"},actor:$a}')" >/dev/null

status="$(curl -fsS "${API}/v1/admin/components/agent/coin-agent/versions/${AGENT_VERSION}" "${AUTH[@]}" | jq -r .status)"
if [[ "${status}" != "draft" ]]; then
  echo "expected draft status, got ${status}" >&2
  exit 1
fi

DELETE_VERSION="${COIN_AGENT_DELETE_E2E_VERSION:-9.9.9-hub-delete}"
echo "==> agent draft delete ${DELETE_VERSION}"
api_post "/v1/admin/components/agent/coin-agent/versions/drafts" \
  "$(jq -n --arg v "${DELETE_VERSION}" --arg a "${ACTOR}" --arg d "${E2E_DIGEST}" \
    '{version:$v, metadata:{image:"nexus:8082/coin-docker/coin-agent:'"${DELETE_VERSION}"'",digest:$d,runtime:"coin-agent"},actor:$a}')" >/dev/null
del_code="$(api_delete "/v1/admin/components/agent/coin-agent/versions/${DELETE_VERSION}?actor=${ACTOR}")"
if [[ "${del_code}" != "204" ]]; then
  echo "expected delete HTTP 204, got ${del_code}" >&2
  exit 1
fi
get_code="$(curl -sS -o /dev/null -w '%{http_code}' "${API}/v1/admin/components/agent/coin-agent/versions/${DELETE_VERSION}" "${AUTH[@]}")"
if [[ "${get_code}" != "404" ]]; then
  echo "expected 404 after delete, got HTTP ${get_code}" >&2
  exit 1
fi

echo "==> agent promote ${AGENT_VERSION}"
api_post "/v1/admin/components/agent/coin-agent/versions/${AGENT_VERSION}/promote" \
  "$(jq -n --arg a "${ACTOR}" '{actor:$a}')" >/dev/null
status="$(curl -fsS "${API}/v1/admin/components/agent/coin-agent/versions/${AGENT_VERSION}" "${AUTH[@]}" | jq -r .status)"
if [[ "${status}" != "published" ]]; then
  echo "expected published status, got ${status}" >&2
  exit 1
fi

echo "==> branching-model profile + draft"
api_post "/v1/admin/components" \
  "$(jq -n --arg n "${BML_NAME}" --arg a "${ACTOR}" '{type:"branching-model",name:$n,actor:$a}')" >/dev/null
api_post "/v1/admin/components/branching-model/${BML_NAME}/versions/drafts" \
  "$(jq -n --arg v "${BML_VERSION}" --arg a "${ACTOR}" '{version:$v, actor:$a}')" >/dev/null
bml_status="$(curl -fsS "${API}/v1/admin/components/branching-model/${BML_NAME}/versions/${BML_VERSION}" "${AUTH[@]}" | jq -r .status)"
if [[ "${bml_status}" != "draft" ]]; then
  echo "expected bml draft, got ${bml_status}" >&2
  exit 1
fi

echo "==> branching-model draft delete ${BML_VERSION}"
del_code="$(api_delete "/v1/admin/components/branching-model/${BML_NAME}/versions/${BML_VERSION}?actor=${ACTOR}")"
if [[ "${del_code}" != "204" ]]; then
  echo "expected bml delete HTTP 204, got ${del_code}" >&2
  exit 1
fi
get_code="$(curl -sS -o /dev/null -w '%{http_code}' "${API}/v1/admin/components/branching-model/${BML_NAME}/versions/${BML_VERSION}" "${AUTH[@]}")"
if [[ "${get_code}" != "404" ]]; then
  echo "expected 404 after bml delete, got HTTP ${get_code}" >&2
  exit 1
fi

echo "==> platform component hub E2E OK (agent draft→published, bml draft delete)"
