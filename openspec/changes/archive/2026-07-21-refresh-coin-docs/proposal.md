## Why

Документация в `coin/docs/` расходится с каноном OpenSpec (`openspec/specs/`): всё ещё фигурируют three-pin / `gp-content` pin, Component Studio, `coin-gp-content/`, устаревшие seed-команды. Параллельно не описано ясно **разделение workspace**: sibling-репозитории (`coin-api`, `coin-executor`, `coin-lib`, `coin-ui`) vs meta-репо `coin/` (docs, openspec, docker, starters, samples) vs corp split (P4-03). Нужен единый проход: docs ↔ specs, без dual narrative.

## What Changes

- **Актуализировать** top-level docs (`architecture`, `control-plane`, `golden-paths`, `responsibilities`, `agent-build-model`, `jenkins-setup`, `README`, how-to, runbooks) строго по `openspec/specs/` (прежде всего `gp-release-two-pin`, `gp-embedded-pipeline`, `jenkins-lib-boundary`, `platform-*`, `ui-enabling-shell`).
- **Исправить** спеку `runtime-documentation`: убрать требование three-pin; зафиксировать **2 pin** (`agent`, `branching-model`) + embedded pipeline.
- **Добавить / переписать** документ(ы) про **разделение монорепы / workspace layout**:
  - текущий local pilot layout (sibling checkouts + `coin/` meta);
  - что живёт в `coin/` vs prod-компонентах;
  - удалённые деревья (`coin-gp-content`, `coin-branching-models`, `coin-jenkins-agents`);
  - corp gate P4-03 (`runbooks/prod-repo-split.md`) — расширить и связать с architecture.
- **Пометить superseded** устаревшие ADR-баннеры / секции (исторический контекст оставить, live SoT — нет).
- **Индекс** `docs/README.md`: убрать ссылки на мёртвые пути; reading order от specs-канона.

## Non-goals

- Реализация `pipeline-tekton-alignment` (v4) или смена schema pipeline.
- Удаление UI `/platform/build-stacks` / мёртвого `gp-content` editor в coin-ui.
- Фактический corp `git filter-repo` / prod deploy.
- Переписывание archive OpenSpec changes.
- Полный rewrite каждого исторического ADR в «только настоящее» (достаточно banner + pointer).

## Capabilities

### New Capabilities

- `docs-monorepo-layout`: каноническое описание workspace / monorepo split — sibling repos, роль `coin/`, local pilot vs corp P4-03, inventory removed paths.

### Modified Capabilities

- `runtime-documentation`: three-pin → two-pin + embedded pipeline; docs MUST align with `gp-release-two-pin` / `gp-embedded-pipeline`; cross-link monorepo layout doc.
- `gp-composition-two-slot`: усилить требование согласованности narrative docs (architecture / control-plane / golden-paths) с two-pin (если ещё дыры после refresh).

## Impact

| Область | Что |
|---------|-----|
| **coin/docs/** | массовый refresh top-level + how-to + runbooks + ADR banners |
| **openspec/specs/** | fix `runtime-documentation`; new `docs-monorepo-layout` |
| **docs/runbooks/prod-repo-split.md** | расширить: полный inventory sibling ↔ corp |
| **coin/README.md** | layout workspace согласован с docs |

Канон: [gp-release-two-pin](../../openspec/specs/gp-release-two-pin/spec.md), [gp-embedded-pipeline](../../openspec/specs/gp-embedded-pipeline/spec.md), [jenkins-lib-boundary](../../openspec/specs/jenkins-lib-boundary/spec.md), ADR [gp-embedded-pipeline](../../docs/adr/gp-embedded-pipeline.md), [coin-ci-runtime](../../docs/adr/coin-ci-runtime.md).
