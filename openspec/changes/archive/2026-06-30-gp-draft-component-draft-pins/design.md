## Context

Три слоя уже согласованы в спеках, но UI create path отстаёт:

| Слой | GP draft + draft gc/bm | Promote GP при draft pin |
|------|------------------------|---------------------------|
| **Spec** (`gp-composition-two-slot`, `gp-publish-flows`) | ✅ | ✅ |
| **coin-api** (`componentResolveModeForGPDraftEdit`, `validateGPCompositionPublishedPins`) | ✅ | ✅ 409 |
| **coin-ui edit** (`useGpCompositionEditor`, `GpReleaseDetail` promote disabled) | ✅ | ✅ |
| **coin-ui create** (`PublishWizard`) | ❌ published-only для всех | N/A |

Проблема из explore: профили `*-30-06` с draft versions → wizard пустой.

## Goals / Non-Goals

**Goals:**

- New GP draft wizard: selectable draft versions для `gp-content` и `branching-model`.
- Agent slot: published versions only + понятный hint.
- Wizard показывает draft pin warning (promote GP заблокирован до publish всех pins).
- DRY: один helper для version lists (wizard + editor).
- E2E + docs.

**Non-Goals:**

- Draft agent pin.
- Backend API changes.
- Изменение canary / resolve semantics.

## Decisions

### D1. Version list rules (unchanged product model)

```
agent           → published only
gp-content      → draft + published
branching-model → draft + published
```

Соответствует `componentResolveModeForGPDraftEdit` и `gp-composition-two-slot`.

### D2. Shared helper

Вынести из `useGpCompositionEditor.ts`:

```ts
versionLabels(items, publishedOnly: boolean): string[]
publishedVersions(items): string[]
```

в `coin-ui/src/lib/gpCompositionVersions.ts` (или `gpSlots.ts`). Wizard и editor импортируют один источник.

### D3. PublishWizard changes

| Location | Change |
|----------|--------|
| Initial load + `useEffect` per slot | gc/bm: `versionLabels(..., false)`; agent: `publishedVersions` |
| `versionStatuses` state | Добавить в wizard (как editor) для `GpCompositionForm` badges |
| `draftBlocked` | Разрешить submit при draft gc/bm pins; блокировать только если нет версий вообще или нет published agent |
| Empty option text | Различать «нет published» (agent) vs «нет версий» (gc/bm) |

`GpCompositionForm` уже рендерит `(draft)` suffix и warning panel — передать `versionStatuses`.

### D4. Promote gate (no new work)

`GpReleaseDetail`: `disabled={draftPinCount > 0}` + API `409` + blocking pins list — оставить как есть.

Wizard: informational warning only (no Promote button on create screen).

### D5. E2E

Extend `e2e-gp-draft-canary-resolve.sh` или добавить UI-agnostic step в новый script:
1. Create gp-content draft
2. Create GP draft via API with draft gc pin (already in e2e-gp-draft-canary-resolve)
3. Assert promote returns 409 with blocking pins
4. Optional: document manual UI verification checklist

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| User ожидает draft agent | Hint на agent slot + link to runtime hub |
| Дублирование логики wizard vs editor | Shared helper (D2) |
| GP draft с незавершёнными component drafts | Canary resolve OK; promote blocked — by design |

## Migration Plan

1. Deploy coin-ui only.
2. No data migration.

Rollback: revert wizard filters.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Draft agent в GP? | ✅ | Нет — published only (существующий контракт) |
| Q2 | Менять API? | ✅ | Нет — UI gap only |
