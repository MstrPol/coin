# Design: GP Component Platform (UI-first)

## Context

Три слоя SoT (инвариант Coin): Content (Nexus), Metadata (PG), Runtime cache (Nexus pointers). Canary уже реализован: `catalog.latest` / `latest_canary`, `project.canary_mode`, `X-Coin-Channel` ([docs/canary.md](../../docs/canary.md)).

Проблема — отсутствие единого контракта между слоями per component type и authoring вне UI.

## Goals / Non-Goals

**Goals:** UI-first lifecycle; Component Package Model; generic resolve; Component Studio MVP (branching-model green field в GCP-1).

**Non-goals:** corp fleet; OIDC prod; полный gp-content migration до GCP-3.

## Decisions

### D1: UI-first authoring

Enabling team создаёт и выпускает platform components только через coin-ui → Admin API → Nexus. Git/Gitea — optional export, не critical path.

### D2: Lifecycle states

| State | Resolve (stable channel) | Resolve (canary channel) |
|-------|--------------------------|---------------------------|
| draft | ❌ | ❌ |
| canary | ❌ (кроме pilot / canary GP) | ✅ pilot projects |
| published | ✅ | ✅ |

### D3: Component Package Model

- Nexus: `maven-releases/coin/{type}/{name}/{version}/package.manifest.json` + artifacts
- PG `component_versions`: package URL, digest, `content_ref` v2 (без больших тел)
- Resolve: slot registry + materializers → денормализованный manifest snapshot

### D4: Promote pipeline

Единый UI flow после health gate (build reports): component `canary` → `published`, catalog `latest_canary` → `latest`, GP stable release.

### D5: ADR placement

Постоянные решения — [`docs/adr/gp-component-package-model.md`](../../docs/adr/gp-component-package-model.md) (GCP-0.1 ✅).

## Migration (strangler)

| Фаза | Scope |
|------|-------|
| GCP-0 | ADR + lifecycle API contract + inventory |
| GCP-1 | Component Studio + branching-model E2E |
| GCP-2 | GP promote wizard |
| GCP-3 | gp-content в Studio |
| GCP-4 | lib manifest + Nexus HTTP |
| GCP-5 | Fleet cleanup (seed, dual-write) |

## Open Questions (✅ platform lead — 2026-06-23)

| # | Решение |
|---|---------|
| Q1 | PG bodies только draft; published Nexus-only |
| Q2 | lib section + zip ref |
| Q3 | PG metadata + Docker registry |
| Q4 | GCP-0+1 ∥ branching green field |

Q5/Q6: UI-first; Gitea product samples only.

## Risks

- Дублирование legacy plan и OpenSpec artifacts — legacy plan → read-only после миграции tasks
- Scope creep в Component Studio — ограничить GCP-1 UI-01,03,04,06
