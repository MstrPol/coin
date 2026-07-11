## Context

**Текущее состояние:**

- GP hub (`/gp/{name}`) — эталон: Overview, Releases, Policy, Canary; primary action «New draft»; release detail под hub URL.
- Platform каталоги — **плоские** списки всех версий всех компонентов типа; нет профильного hub.
- `PlatformCatalogPage`: create только для `gp-content` («Create draft»); branching-models и runtime — без create.
- `platform-component-lifecycle` spec: runtime **published only**, без draft UI.
- `publish-agent.sh`: `POST /v1/admin/components/agent/coin-agent/versions` → **сразу `published`** (минуя draft).
- coin-api уже имеет generic endpoints: `POST .../versions/drafts`, `POST .../promote`; `componentResolveModeForGPDraftEdit("agent")` = stable-only.
- Executor — derived pin: `executorPinForAgentStack("coin-agent", version)` → `executor/coin-executor@{version}`.

**Prerequisite:** `platform-native-lifecycle` ✅ archived.

**Stakeholders:** enabling team (operator console), platform team (agent image CI).

## Goals / Non-Goals

**Goals:**

- Единый **Platform Component Hub** pattern для agent, gp-content, branching-model (по образцу `gp-entity-hub`).
- Каталог семейства → список **профилей** (имён); hub → Overview + Releases + release detail.
- Унифицировать UX: «New profile» на каталоге, «New draft» на hub.
- Agent stack: полный lifecycle **draft → published** в registry; Jenkins register draft после push image; UI promote + manual catch-up.
- Release detail agent: metadata (image, digest, goarch) + read-only derived executor line.
- Redirects со старых flat URLs на hub hierarchy.

**Non-Goals:**

- Docker/image build UI в coin-ui.
- Отдельный hub/lifecycle для `executor`.
- Draft agent pins в GP composition (см. Open Questions — default: **не** в этом change).
- Corp fleet rollout.

## Decisions

### D1: Shared hub layout component

**Решение:** `PlatformComponentHubLayout` — параметризованный аналог `GpHubLayout`:

| Параметр | agent | gp-content | branching-model |
|----------|-------|------------|-----------------|
| family segment | `runtime` | `build-stacks` | `branching-models` |
| profile label | Agent stack | Build stack | Branching model |
| tabs | Overview, Releases | Overview, Releases | Overview, Releases |
| draft editor | metadata form (no artifacts) | `PlatformComponentEditor` | `PlatformComponentEditor` |

**Альтернатива:** три отдельных layout — отклонено (дублирование).

**URL map:**

```
/platform/runtime                          → catalog (profiles)
/platform/runtime/new                      → create agent profile
/platform/runtime/{name}                   → hub Overview
/platform/runtime/{name}/releases          → Releases tab
/platform/runtime/{name}/releases/new-draft → new draft (version + optional metadata catch-up)
/platform/runtime/{name}/{version}         → release detail
/platform/runtime/{name}/{version}/edit    → draft metadata edit (publisher)

/platform/build-stacks/{name}              → hub (аналогично)
/platform/branching-models/{name}          → hub (аналогично)
```

### D2: Catalog shows profiles, not version rows

**Решение:** каталог группирует `components` API по `name`; колонки: name, latest published, draft count, updated. Клик по имени → hub.

**Альтернатива:** оставить flat version list — отклонено (не соответствует GP mental model).

### D3: Agent draft lifecycle in API

**Решение:**

1. `publish-agent.sh` → `POST .../versions/drafts` с metadata `{image, digest, goarch, runtime}`.
2. После smoke/approval → `POST .../versions/{version}/promote` (или отдельный admin promote endpoint).
3. Идемпотентность: повторный POST drafts с тем же version → `409` (как сейчас); Jenkins трактует `409` как success.
4. Promote для agent **не требует** Nexus content_ref (в отличие от gp-content); достаточно metadata с image ref. Store: skip `validateContentRefOnWrite` для `type=agent` on draft insert.

**Альтернатива:** оставить direct publish — отклонено (нет draft gate, нет UI catch-up).

**Migration:** существующие `published` agent versions без изменений; новые версии — через draft path.

### D4: Executor as derived read-only line

**Решение:** hub release detail для agent показывает секцию «Derived executor pin»: `executor/coin-executor@{same version}` — вычисляется клиентом или из API field `derivedExecutorPin`. Не создавать отдельную строку в Releases tab.

**Альтернатива:** отдельный executor hub — отклонено (superseded by agent stack model).

### D5: Unified action labels

**Решение:**

| Context | Label |
|---------|-------|
| Family catalog (publisher) | «New profile» |
| Profile hub (publisher) | «New draft» |
| Rename «Create draft» → «New draft» everywhere |

### D6: Redirects

**Решение:**

| Legacy | Target |
|--------|--------|
| `/platform/build-stacks/{name}/{version}` | `/platform/build-stacks/{name}/releases/{version}` or keep short URL with hub breadcrumb |
| `/platform/build-stacks/{name}/{version}/edit` | unchanged (editor) |
| `/components/agent/{name}` | `/platform/runtime/{name}` |

Предпочтение: **короткий** `/{name}/{version}` под hub layout (как GP `/gp/{name}/releases/{version}`) — унифицировать на `.../releases/{version}` для всех трёх families.

### D7: GP composition links

**Решение:** GP release detail composition links → `/platform/runtime/{agentName}/releases/{version}` (и аналоги для gp-content, branching-model).

**GP draft agent pin rule (default this change):** без изменений — только `published` agent в composition picker и API validation. Draft agent существует в registry, но не pin'ится в GP draft до promote.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking bookmarks на flat catalog URLs | Redirects + hub breadcrumbs |
| Jenkins публикует образ, но draft register fails | Runbook: retry register; UI catch-up form на hub |
| Agent promote без content_ref ломает generic promote validation | Type-specific promote path для `agent` |
| Три hub implementation drift от GP | Shared `PlatformComponentHubLayout` + tab components |
| Operator путает agent profile и GP profile | Чёткие labels: «Agent stack» vs «Golden Path profile» |

## Migration Plan

1. **coin-api:** allow agent draft insert/promote; optional `derivedExecutorPin` in version GET response.
2. **coin-ui:** hub routes + refactor catalogs; keep old routes as redirects one release.
3. **publish-agent.sh:** switch to drafts endpoint + promote step (promote может быть отдельным manual gate до corp).
4. **E2E:** update navigation asserts; add agent draft register + promote flow.
5. **Rollback:** revert UI routes; Jenkins может временно вернуть direct publish endpoint (backward compat alias).

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Разрешить **draft agent pins** в GP draft composition (симметрия с gp-content)? | ✅ | **A:** published-only в GP draft (default) / **B:** draft+publisher в picker | **A** — agent pin в GP draft/composition остаётся `published` only; draft agent живёт в registry до promote |
| Q2 | Auto-promote agent после Jenkins register или manual promote в UI? | ✅ | **A:** Jenkins вызывает promote сразу после draft / **B:** manual promote only | **A** для CI (`publish-agent.sh`); **B** для UI catch-up/backfill |
