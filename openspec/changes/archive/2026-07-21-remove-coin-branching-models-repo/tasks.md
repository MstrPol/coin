## 1. Fixtures и schema

- [x] 1.1 Скопировать `models/trunk-based/model.yaml` и `models/semver-tag/model.yaml` в `docker/testdata/branching-models/<name>/model.yaml`
- [x] 1.2 Перенести `schemas/branching-model.schema.json` в `docs/schemas/branching-model.schema.json`

## 2. Seed и E2E

- [x] 2.1 Добавить `docker/scripts/seed-branching-model.sh` (логика publish из старого `publish-branching-model.sh`, путь к testdata)
- [x] 2.2 Обновить `docker/scripts/seed-jenkins-lib-stack.sh`: вызывать seed-helper вместо `coin-branching-models/scripts/...`
- [x] 2.3 Обновить `e2e-branching-model-lifecycle.sh`, `e2e-branching-canary-resolve.sh`, `e2e-gp-promote-gate.sh`: `MODEL_YAML` → testdata

## 3. Docs и ADR

- [x] 3.1 Обновить `docs/how-to/branching-models.md`: убрать ссылки на `coin-branching-models/`; добавить schema + testdata
- [x] 3.2 Обновить ADR `docs/adr/gp-branching-model.md`: убрать Reference catalog на git-дерево; зафиксировать Platform SoT + testdata для pilot seed
- [x] 3.3 Синхронизировать main specs `openspec/specs/branching-models-catalog` и `openspec/specs/branching-model` с delta (или оставить на archive — по процессу sync)

## 4. Удаление каталога

- [x] 4.1 Удалить дерево `coin-branching-models/`
- [x] 4.2 `rg coin-branching-models` по активному дереву `coin/` (исключая `openspec/changes/archive`) — ноль совпадений
- [x] 4.3 При доступном local stack: прогнать seed / релевантные branching E2E
