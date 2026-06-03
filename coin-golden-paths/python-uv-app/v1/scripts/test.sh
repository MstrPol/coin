#!/usr/bin/env bash
set -euo pipefail

echo "==> coin standard test (python-uv)"
uv sync --frozen --all-groups
uv run pytest
