#!/usr/bin/env bash
set -euo pipefail

echo "==> coin standard test (python-pip)"
python -m venv .venv
. .venv/bin/activate
python -m pip install --upgrade pip
python -m pip install -r requirements.txt
if [[ -f requirements-dev.txt ]]; then
  python -m pip install -r requirements-dev.txt
fi
export PYTHONPATH="${PWD}/src:${PYTHONPATH:-}"
python -m pytest
