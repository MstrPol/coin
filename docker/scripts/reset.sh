#!/usr/bin/env bash
# Остановить стенд и удалить volumes (полный сброс данных).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"
docker compose down -v --remove-orphans
rm -f "${ROOT}/.nexus-admin-initialized"
echo "Coin local stack removed (volumes удалены). Подъём с нуля: make bootstrap"
