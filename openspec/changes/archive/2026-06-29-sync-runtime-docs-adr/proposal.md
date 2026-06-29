## Why

После hard cut build-engine model, branching schema v2 и three-pin GP composition документация и ADR расходятся с кодом: superseded ADR без явных шапок, `architecture.md` описывает 4-slot composition с отдельным executor, примеры manifest всё ещё содержат `publish when: tag`, а operational truth размазана между `agent-build-model.md`, `control-plane.md` и ADR с legacy-контекстом.

Перед работой над gp-content schema v2 / build stack editor нужен единый SoT по runtime-модели в `docs/`, чтобы proposal и ADR не противоречили executor/coin-lib.

## What Changes

- Добавить ADR **`coin-ci-runtime`** — каноническое описание: coin-agent, bootstrap, 3 build engines, pilot (podman) vs corp (buildkitd), publish gate (branching + `params.publish`), GP composition 3 pins.
- Актуализировать **`build-engine-contract.md`**: явно отделить legacy-контекст от текущего состояния; ссылка на `coin-ci-runtime` как operational SoT.
- Пометить superseded ADR в **теле файлов** (`gp-composition-four-components`, `gp-pipeline-bundle-layer`): статус, дата, замена.
- Обновить **`docs/adr/README.md`**: `gp-branching-model`, `coin-ci-runtime`; актуальный индекс.
- Синхронизировать **`architecture.md`**, **`control-plane.md`**, **`agent-build-model.md`**, **`docs/README.md`**: three-pin composition, derived executor, branching в manifest, publish без `when: tag` в примерах.
- Убрать/пометить устаревшие ссылки на pipeline-bundle, jnlp slot, stack agents, `when: tag` как primary publish gate.
- **Non-goals:** gp-content schema v2, build stack visual editor, preview API, изменения runtime-кода (coin-executor, coin-lib, coin-api), corp migration rollout.

## Capabilities

### New Capabilities

- `runtime-documentation`: требования к согласованности ADR и docs с фактической CI runtime-моделью Coin.

### Modified Capabilities

- `build-engine`: уточнить doc-level требования (pilot podman implementation vs engine name `buildkit`; bootstrap без buildkitd на arm64 pilot).
- `gp-composition-two-slot`: синхронизировать narrative docs с three-pin composition (spec уже верный; delta — doc traceability / cross-links).

## Impact

- `docs/adr/*.md`, `docs/architecture.md`, `docs/control-plane.md`, `docs/agent-build-model.md`, `docs/README.md`
- Возможные правки `coin-ui/README.md`, `docs/how-to/*.md` при обнаружении legacy-формулировок
- Без изменений API, manifest schema, executor behavior
- Разблокирует change `gp-content-schema-v2` (build stack) на согласованной doc-базе
