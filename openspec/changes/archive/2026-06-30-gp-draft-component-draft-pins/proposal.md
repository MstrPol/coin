## Why

Enabling team собирает GP из экспериментальных platform profiles (`agent-30-06`, `bs-30-06`, `bm-30-06`), но **New GP draft wizard** (`PublishWizard`) показывает «— нет версий —», потому что фильтрует **все** слоты до `published`. При этом спека и coin-api уже разрешают GP draft с draft pins для `gp-content` и `branching-model`, а promote GP блокируется, пока хотя бы один pin не `published`.

Нужно закрыть разрыв UI ↔ контракт: создавать GP draft на draft build stack / branching model, но **не** выпускать stable GP release (promote), пока в composition есть draft component.

## What Changes

- **coin-ui `PublishWizard`:** version pickers для `gp-content` и `branching-model` — `draft` + `published` (как `useGpCompositionEditor`); `agent` — **только published** (без изменения правила).
- **coin-ui:** статус pin `(draft)` / badge + предупреждение «Draft pins блокируют promote GP» в wizard (reuse `GpCompositionForm` behavior).
- **coin-ui:** явные empty states — «Нет published agent» vs «Нет версий — создайте draft на Platform hub».
- **Shared helper:** вынести `versionLabels` / `publishedVersions` в общий модуль (wizard + composition editor).
- **coin-api:** без изменений контракта (create GP draft с draft gc/bm уже работает; promote 409 на draft pins уже есть).
- **E2E:** сценарий создания GP draft через API/UI с draft gp-content pin + verify promote blocked.
- **Docs:** `docs/coin-ui-user-guide.md` — workflow «GP draft на component drafts → publish pins → promote GP».

### Non-goals

- Draft **agent** pin в GP composition (остаётся published-only — CI runtime stability).
- Авто-publish component drafts при promote GP.
- Изменение canary resolve rules (draft pins на canary уже допустимы).
- Build stack visual editor / yaml schema work.
- Corp fleet rollout.

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `gp-publish-flows`: wizard creation MUST expose draft versions для gp-content/branching-model; promote gate UX в create flow.
- `gp-composition-two-slot`: уточнить end-to-end сценарий «GP draft на draft component pins» vs promote gate.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-ui** | `PublishWizard.tsx`, shared version helper, optional `GpCompositionForm` props |
| **coin-api** | verify only (no API change expected) |
| **docker/e2e** | GP draft + draft pin + promote gate smoke |
| **docs** | coin-ui user guide — GP draft composition workflow |
