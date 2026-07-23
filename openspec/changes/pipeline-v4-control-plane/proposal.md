## Why

Контракт `schemaVersion: 4` и file-resolve петля уже стабилизированы в change `pipeline-tekton-alignment` (executor + coin-lib + fixture без live control plane). Теперь нужно, чтобы **coin-api materialize’ил тот же shape**, UI редактировал tasks/containerfiles, а product repos ходили через **remote resolve** с полным E2E.

## What Changes

- coin-api: validate/storage schemaVersion 4 на GP release body; manifest builder materialize `containerfiles[]` + `pipeline.tasks` + `destinations[]` (= fixture shape из alignment).
- coin-api: preview API v4; v3→v4 migration on draft save + v3 read adapter на переходный период.
- Seed pipelines `go-app` / `go-app-docker` v4; reseed PG/Nexus.
- coin-ui: containerfiles catalog panel; task graph editor (`tasks`, `runAfter`, steps); layout Composition → Pipeline → Containerfiles → Parameters; preview + migration UX.
- Remote E2E: `demo-go-app` / `demo-go-app-docker` через `coin.resolve: remote` (не file).
- Sync/archive delta specs в main после green E2E.

Depends on: `pipeline-tekton-alignment` (resolved shape SoT для builder).

## Capabilities

### New Capabilities

- `gp-v4-materialize`: coin-api builder/preview/validate v4 resolved shape (tasks + containerfiles + destinations catalog).
- `gp-v4-migration`: v3→v4 draft migration + temporary v3 read adapter.
- `pipeline-v4-ui`: UI editors для catalog/tasks/preview.

### Modified Capabilities

- `manifest-pipeline-inline`: remote materialize MUST match file-fixture shape из alignment.
- `gp-embedded-pipeline`: seed/storage schemaVersion 4 для go-app samples.

## Non-goals

- Повторная отладка executor без API (это AC `pipeline-tekton-alignment`).
- Замена Jenkins на Tekton Controller.
- Hard guardrail на `resolve: file` в product repos.
- Corp fleet migration.

## Impact

| Область | Изменение |
|---------|-----------|
| **coin-api** | validate, storage, builder, preview, migrate, seed |
| **coin-ui** | catalog + task graph + preview |
| **samples** | remote resolve E2E (после seed) |
| **openspec** | sync main specs; archive alignment deltas при необходимости |
| **coin-executor / coin-lib** | без новых контрактов — потребляют тот же resolved shape |
