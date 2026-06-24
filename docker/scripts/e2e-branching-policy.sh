#!/usr/bin/env bash
# GBM-3.4: trunk-based publish policy via coin-executor (no Jenkins).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "${ROOT}/.." && pwd)"
EXEC="${REPO_ROOT}/coin-executor/coin-executor"

need() { command -v "$1" >/dev/null || { echo "missing: $1" >&2; exit 1; }; }
need go

echo "==> build coin-executor"
(cd "${REPO_ROOT}/coin-executor" && go build -o coin-executor ./cmd/coin-executor)

work="$(mktemp -d)"

mkdir -p "${work}/.coin"
cat >"${work}/.coin/config.yaml" <<'YAML'
coin:
  goldenPath: go-app
  version: "1.0.0"
jenkins:
  credentials:
    docker: nexus-docker
project:
  name: demo-go-app
  artifactId: demo-go-app
  groupId: com.example
  repository: maven-releases
deliverables:
  app:
    type: image
YAML

cat >"${work}/.coin/manifest.json" <<'JSON'
{
  "manifestVersion": 1,
  "goldenPath": {"name": "go-app", "version": "1.0.0"},
  "executor": {"version": "0.1.0", "url": "http://example/executor"},
  "runtime": {"image": "coin-agent:1.0.0"},
  "build": {
    "engine": "buildkit",
    "buildkit": {
      "dockerfile": ".coin/Containerfile",
      "targets": {"validate": "validate", "test": "test", "image": "runtime"},
      "containerfile": {"url": "http://example/containerfile", "sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
    }
  },
  "pipeline": {
    "stages": [
      {"id": "validate", "name": "Validate"},
      {"id": "publish", "name": "Publish", "when": "tag"}
    ]
  },
  "validateSchema": {"url": "http://example/schema", "sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
  "capabilities": {"deliverables": ["image"]},
  "branching": {
    "name": "trunk-based",
    "version": "1.0.0",
    "trunk": {"branch": "main"},
    "branchTypes": ["feature", "bugfix", "release"],
    "versioning": {
      "tagPrefix": "v",
      "qualifiers": {
        "snapshot": {"enabled": true},
        "rc": {"enabled": true, "releaseBranchesOnly": true}
      }
    },
    "publish": {"when": "tag"}
  }
}
JSON

echo "==> mock coin-api policy-check"
policy_port="$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()')"
python3 - "${policy_port}" <<'PY' &
import json, sys
from http.server import BaseHTTPRequestHandler, HTTPServer

port = int(sys.argv[1])

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps({"warning": ""}).encode())

    def log_message(self, *_):
        return

HTTPServer(("127.0.0.1", port), Handler).serve_forever()
PY
policy_pid=$!
trap 'kill "${policy_pid}" 2>/dev/null || true; rm -rf "${work}"' EXIT
export COIN_API_URL="http://127.0.0.1:${policy_port}"
export COIN_API_TOKEN=test-token

echo "==> validate on feature branch (expect pass)"
(
  cd "${work}"
  export GIT_BRANCH=feature/PROJ-101
  unset TAG_NAME GIT_TAG_NAME
  "${EXEC}" validate --project .coin/config.yaml --manifest .coin/manifest.json
)

echo "==> publish on feature branch without tag (expect skip exit 0)"
out="$(
  cd "${work}"
  export GIT_BRANCH=feature/PROJ-101
  unset TAG_NAME GIT_TAG_NAME
  "${EXEC}" run --project .coin/config.yaml --manifest .coin/manifest.json --stage publish 2>&1
)"
echo "${out}"
if ! grep -q "skip stage publish" <<<"${out}"; then
  echo "FAIL: expected publish skip on feature branch" >&2
  exit 1
fi

echo "OK: trunk-based branching policy E2E"
