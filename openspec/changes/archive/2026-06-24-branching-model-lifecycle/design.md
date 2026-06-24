## Context

**Текущее состояние:**

- UI: branching-model виден только в `/components` без доменного каталога; Studio на publish canary вызывает `register-package` → Nexus upload + `content_ref` v2 с `package.url`.
- API: `RegisterComponentPackage` всегда пишет в Nexus ([`coin-api/internal/admin/component_package.go`](../../../coin-api/internal/admin/component_package.go)).
- ADR Q1 ([`docs/adr/gp-component-package-model.md`](../../../docs/adr/gp-component-package-model.md)): PG bodies для draft Studio; published — Nexus-only. Canary фактически обошёл это правило.

**Stakeholders:** enabling team (authoring), platform lead (promote gate), product CI (resolve stable + canary).

## Goals / Non-Goals

**Goals:**

- Отдельный UI-каталог branching models с полным lifecycle view.
- Canary = PG SoT + resolve через API; Nexus immutable package только на `published`.
- Сохранить Nexus fallback для stable/published pins.
- Первый тип на новом lifecycle: `branching-model` (остальные types — follow-up с тем же API).

**Non-Goals:**

- Миграция всех gp-content versions в этом change.
- Fleet corp rollout.
- Удаление уже загруженных в Nexus canary packages (cleanup runbook отдельно).

## Decisions

### D1: content_ref v2 — два фазы

| Phase | status | content_ref | Resolve source |
|-------|--------|-------------|----------------|
| Draft / Canary | `draft`, `canary` | `{ schemaVersion: 2, manifest: { branching: {...} } }` — **без** `package` | PG `component_artifact_bodies` + manifest subset |
| Published | `published` | полный v2 с `package.url` + `package.sha256` + `manifest` | Nexus (fallback) + registry |

**Альтернатива A:** canary с Nexus staging repo — отклонено (immutable repos, риск «утечки» нестабильного в fallback).

**Альтернатива B:** отдельный `content_ref` schema v3 — отклонено (лишний blast radius); optional `package` в v2 достаточно.

### D2: API split — `register-draft` vs `publish-to-nexus`

- Переименовать семантику `POST .../register-package`:
  - **Canary path:** `register-package` — validate, build manifest subset, **UpdateContentRef (PG-only)**, **no Nexus**.
  - **Promote path:** `promote` — upload Nexus, финальный content_ref, status `published`.
- `publish-canary` остаётся сменой status; требует предварительный PG content_ref (после register).

### D3: UI — `/branching-models`

- Список: name, latest per channel (`draft`/`canary`/`published`), GP profiles using model, actions (Open Studio, Promote).
- API: `GET /v1/admin/components/branching-model` + enrich с profile slots из `gp_profiles`.
- Не дублировать полный Component Studio — deep link в `/studio/branching-model/{name}/{version}`.

### D4: Resolve materializer

`loadBranchingBundle` для `ComponentResolveCanary` / admin: если `content_ref` без `package` — читать rules из `manifest` subset + optional `model.yaml` body из PG.

Stable resolve для `published`: prefer Nexus package; PG bodies не required.

### D5: Bootstrap scripts

`publish-branching-model.sh` — после API change: draft → register (PG) → canary → promote (Nexus). Для local seed допустим shortcut `promote` сразу после register (skip pilot) — documented in runbook.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| CI fallback не видит canary | By design; canary требует coin-api. Stable fallback unchanged. |
| Старые canary versions уже в Nexus | Orphan blobs; cleanup script / ignore (immutable) |
| content_ref validation ломается без package | Relax `ValidateContentRefV2` — `package` optional until published |
| gp-content всё ещё upload на canary | Document as known gap; follow-up change |

## Migration Plan

1. API: optional package in v2 + register без Nexus + promote с Nexus.
2. UI: каталог + Studio flow update.
3. Update `branching-model` materializer + tests.
4. Fix bootstrap script.
5. Manual E2E: draft → canary → resolve canary (no Nexus) → promote → resolve stable (Nexus).

**Rollback:** revert API; published versions in Nexus remain valid.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Применить Nexus gate ко всем Studio types сразу или только branching-model? | ✅ BML-0.1 | **A:** только `branching-model` в этом change |
| Q2 | Удалять PG bodies после promote? | ✅ BML-0.1 | **A:** keep для admin preview |
