## Context

`pipeline-tekton-alignment` зафиксировал resolved shape v4 (file resolve + executor + coin-lib) и fixture `samples/demo-go-app/.coin/manifest.local.yaml` как SoT формы. Control plane (API storage/builder, UI, remote resolve E2E) вынесен сюда, чтобы alignment можно было закрыть по offline acceptance без ожидания UI/API.

Связанные: ADR [pipeline-tekton-mapping.md](../../docs/adr/pipeline-tekton-mapping.md), change `pipeline-tekton-alignment`.

## Goals / Non-Goals

**Goals:**

- coin-api materialize resolved v4 = fixture shape (tasks, containerfiles, destinations catalog).
- Preview + validate schemaVersion 4 на GP draft/release.
- v3→v4 migration на save; temporary v3 read adapter.
- Seed `go-app` / `go-app-docker` v4; reseed.
- coin-ui: catalog + task graph + preview.
- Remote E2E green на samples без `resolve: file`.

**Non-Goals:**

- Изменение executor step semantics (кроме багфиксов совместимости с materialize).
- Offline file-resolve acceptance (AC alignment).
- Tekton Controller / corp fleet.

## Decisions

### D1: Fixture shape = materialize contract

Builder MUST эмитить тот же top-level shape, что file fixture:

- `schemaVersion: 4`
- `pipeline.tasks[]`, `containerfiles[]`, `destinations[]` (catalog)
- без `pipeline.stages`, без top-level `build` / `deliverables` / `capabilities.deliverables`

JSON Schema `pipeline-inline.v4.schema.json` + `manifest.schema.json` — валидация storage и resolve output.

### D2: Migration v3→v4 on draft save

При save draft с v3 body — автоматическая миграция в v4 (stages→tasks, inline containerfile→catalog entry). Read adapter: старые release blobs v3 читаются до reseed; новые writes только v4.

### D3: UI layout

Composition → Pipeline (task graph) → Containerfiles → Parameters. Preview показывает resolved preview JSON той же формы, что Jenkins получит после materialize.

### D4: Remote E2E gate

Acceptance этого change: `demo-go-app` (и docker sample) SUCCESS при `coin.resolve: remote` (default), без локального fixture path как primary.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Drift builder vs fixture | Shared schema + golden compare с `manifest.local.yaml` shape |
| Breaking remote v3 mid-migration | Read adapter + dual seed window |
| UI scope creep | Strict panels из D3; no parallel DSL editor |

## Migration Plan

1. API validate + storage v4.
2. Builder + preview; golden tests vs fixture shape.
3. Migration + seed/reseed.
4. UI editors.
5. Switch samples to remote; E2E.
6. Sync main specs; archive related deltas.

Rollback: leave remote on last green v3 seed; UI feature-flag editors.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | UI graph library | ⏳ | A: custom React / B: React Flow | — platform lead |
| Q2 | Срок жизни v3 read adapter | ⏳ | A: до первого reseed / B: N релизов | — |
