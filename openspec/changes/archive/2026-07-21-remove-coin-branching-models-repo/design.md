## Context

SoT для `branching-model` уже control plane (Platform → PG → Nexus; runtime — `manifest.branching` в coin-executor). Каталог `coin-branching-models/` остался как:

- эталоны `models/*/model.yaml` + README;
- `schemas/branching-model.schema.json` (не используется coin-api — валидация Go в `componentpackage`);
- `scripts/publish-branching-model.sh` (bootstrap seed).

Потребители в `coin`: `seed-jenkins-lib-stack.sh`, `e2e-branching-*.sh`, `e2e-gp-promote-gate.sh`, how-to и ADR.

## Goals / Non-Goals

**Goals:**

- Убрать dual path: один SoT для моделей — Platform/registry.
- Сохранить green local seed и E2E без git-каталога.
- Оставить schema v2 и how-to как единственную human-документацию.

**Non-Goals:**

- Менять GP composition / API / UI lifecycle.
- Fleet/corp rollout.
- Переписывать Go-валидацию на JSON Schema engine.

## Decisions

### D1. Fixtures → `docker/testdata/branching-models/`

| | |
|--|--|
| **Выбор** | Перенести `trunk-based/model.yaml` и `semver-tag/model.yaml` в `docker/testdata/branching-models/<name>/model.yaml`. |
| **Почему** | Seed/E2E уже живут в `docker/scripts`; рядом с ними testdata — стандартный bootstrap path, не «product catalog». |
| **Альтернатива A** | Embed в coin-api (как gpcontent seed) — лишний cross-repo scope; coin-api уже не читает эти файлы для validate. |
| **Альтернатива B** | Inline YAML heredoc в каждом скрипте — дубли и drift. |

Per-model README не переносим: содержимое cheat-sheet уже / должно быть в `docs/how-to/branching-models.md`.

### D2. Schema → `docs/schemas/branching-model.schema.json`

| | |
|--|--|
| **Выбор** | Перенести JSON Schema в docs как контрактную документацию; how-to ссылается на неё. |
| **Почему** | Runtime SoT валидации — Go (`validate_draft.go`) + UI (`branchingModelYaml.ts`). JSON Schema нужна operators/docs, не отдельному репо. |
| **Альтернатива** | Только описание в markdown без `.json` — хуже для diff/IDE; файл дёшев. |

### D3. Seed: тонкий helper в `docker/scripts/`

| | |
|--|--|
| **Выбор** | Вынести логику publish (create component → draft → PUT artifact → register → promote) в `docker/scripts/seed-branching-model.sh`, читающий testdata. `seed-jenkins-lib-stack.sh` вызывает его. |
| **Почему** | Сохраняем идемпотентность seed; убираем зависимость от `coin-branching-models/scripts`. |
| **Альтернатива** | Полностью inline в `seed-jenkins-lib-stack.sh` — раздувает уже длинный скрипт. |

E2E скрипты меняют `MODEL_YAML` path на testdata; логику API не трогаем.

### D4. ADR / docs

Обновить [gp-branching-model.md](../../docs/adr/gp-branching-model.md):

- убрать строку «Reference catalog = coin-branching-models»;
- Authoring = Platform hub; эталоны local pilot = `docker/testdata/...`;
- schema doc path = `docs/schemas/...`.

How-to: примеры inline + ссылка на schema и testdata (для seed), без git-каталога.

## Risks / Trade-offs

| Риск | Митигация |
|------|-----------|
| Seed/E2E сломаются после удаления каталога | Сначала перенести fixtures + обновить скрипты, потом `rm -rf coin-branching-models` |
| Потеря semver-tag эталона | Перенести оба model.yaml в testdata |
| Docs ссылки 404 | Grep по `coin-branching-models` в `coin/` (вне archive) и поправить |
| Кто-то снаружи зовёт publish-скрипт | Non-goal corp; в README/ADR явно: bootstrap только через docker seed |

## Migration Plan

1. Добавить `docker/testdata/branching-models/{trunk-based,semver-tag}/model.yaml` (копия текущих).
2. Добавить `docs/schemas/branching-model.schema.json`.
3. Добавить `docker/scripts/seed-branching-model.sh`; переключить seed + E2E.
4. Обновить how-to + ADR + спеки.
5. Удалить `coin-branching-models/`.
6. Проверка: `rg coin-branching-models` в активном дереве = 0 (кроме archive OpenSpec); прогнать seed/релевантные e2e при возможности.

**Rollback:** восстановить каталог из git history; скрипты откатить. Данные PG/Nexus не затрагиваются.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| — | _(нет blocking)_ | — | — | Seed/testdata + docs/schemas приняты как default для local pilot |
