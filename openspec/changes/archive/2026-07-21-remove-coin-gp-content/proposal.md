## Why

Папка `coin/coin-gp-content/` дублирует bootstrap seed, который уже живёт в `coin-api/internal/gpcontent/seed/` (pipeline YAML, schema, Containerfile). После [gp-embedded-pipeline](../../docs/adr/gp-embedded-pipeline.md) authoring — на GP release; component `gp-content` снят. Папка + `make coin-gp-content` + deprecated `publish-content.sh` создают dual path и путают SoT (как с `coin-branching-models`).

## What Changes

- **Удалить** дерево `coin/coin-gp-content/` (stacks, schemas, scripts, Jenkinsfile, dist, README).
- **Убрать** local glue: `docker/scripts/coin-gp-content.sh`, CASC `casc-coin-gp-content-build.yaml`, `make coin-gp-content`, вызовы из bootstrap/README.
- **Зафиксировать SoT seed:** `coin-api/internal/gpcontent/seed/` — единственный bootstrap source для pilot GP pipeline defaults.
- **Обновить** docs/ADR: убрать ссылки на `coin-gp-content/stacks` как SoT; указать api seed + GP release authoring.
- **Согласовать** активный change `pipeline-tekton-alignment`: tasks про миграцию stacks → retarget на `coin-api/.../seed` (не держать v4 в двух местах).
- Перед удалением: **проверка**, что api seed покрывает pilot stacks (`go-app`, `go-app-docker`) — без регресса `seed-jenkins-lib`.

## Non-goals

- Удаление embedded pipeline / GP release pipeline editor / capability build policy.
- Удаление UI `/platform/build-stacks` и мёртвого `gp-content` editor в coin-ui (отдельный cleanup; migration 033 уже purge'нул type из DB).
- Изменение schema pipeline-inline (v3→v4) — остаётся в `pipeline-tekton-alignment`.
- Правки кода coin-api/coin-executor сверх docs coordination (seed уже на месте; этот change scoped к репо `coin`).
- Corp fleet / prod-repo-split wave (обновить runbook текст, не делать split).

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `gp-embedded-pipeline`: bootstrap seed — `coin-api` embed; папка `coin-gp-content/` не требуется.
- `runtime-documentation`: architecture/control-plane/ADR не ссылаются на `coin-gp-content/` как SoT build content; seed path = api embed.

## Impact

| Область | Что |
|---------|-----|
| **coin** | удаление папки; docker make/bootstrap/CASC; docs + ADR (`gp-embedded-pipeline`, `build-engine-contract`, `coin-ci-runtime`, `architecture`, …) |
| **openspec** | delta specs; правка tasks в `pipeline-tekton-alignment` (retarget seed) |
| **coin-api** | без обязательных code changes (seed уже SoT); verify-only |
| **Local pilot** | `make seed-jenkins-lib` / E2E без `make coin-gp-content` |

Связанные ADR: [gp-embedded-pipeline](../../docs/adr/gp-embedded-pipeline.md), [build-engine-contract](../../docs/adr/build-engine-contract.md), [coin-ci-runtime](../../docs/adr/coin-ci-runtime.md).
