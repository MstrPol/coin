#!/usr/bin/env bash
# Остановить стенд, volumes сохранить.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"
docker compose down --remove-orphans
echo "Coin local stack stopped (volumes сохранены). Запуск: make up"
