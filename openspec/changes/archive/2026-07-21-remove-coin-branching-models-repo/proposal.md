## Why

Каталог `coin-branching-models/` дублирует SoT: authoring и lifecycle моделей ветвления уже живут в Platform UI → coin-api (PG) → Nexus, runtime читает `manifest.branching` в coin-executor. Git-каталог остался как reference + bootstrap seed и создаёт dual path (follow-up из [remove-executor-component](../archive/2026-06-30-remove-executor-component/proposal.md)).

## What Changes

- **Удалить** дерево `coin-branching-models/` (models, schemas, scripts, README, dist).
- **Перенести** эталонные `model.yaml` (как минимум `trunk-based`) в `docker/testdata/branching-models/` для local seed и E2E.
- **Перенести** `branching-model.schema.json` в `docs/schemas/` (контрактная документация; runtime-валидация остаётся в coin-api Go + coin-ui).
- **Обновить** `docker/scripts/seed-jenkins-lib-stack.sh` и E2E-скрипты: публиковать/читать fixtures из testdata, без вызова `publish-branching-model.sh`.
- **Обновить** docs/ADR: убрать «Reference catalog = coin-branching-models»; SoT = Platform + PG/Nexus.
- **Обновить** спеку `branching-models-catalog`: ссылки на how-to, без per-model README в git-каталоге.

## Non-goals

- Удаление component type `branching-model`, GP pin, Platform UI `/platform/branching-models`, или runtime policy в coin-executor.
- Изменение schema v2 (`branches[]`, template DSL) или three-pin composition.
- Corp fleet migration / dual path для старых git-publish workflows вне local pilot.
- Перенос schema в отдельный репозиторий coin-api (валидация уже Go-native; JSON schema — docs в `coin`).

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `branching-models-catalog`: убрать требование ссылок на `coin-branching-models/models/{name}/README`; оставить how-to.
- `branching-model`: уточнить, что эталоны/schema живут в docs/testdata control plane, не в отдельном git-каталоге reference models.

## Impact

| Область | Что затронуто |
|---------|----------------|
| **coin** | удаление `coin-branching-models/`; `docker/scripts/*`; `docker/testdata/`; `docs/how-to/branching-models.md`; ADR `gp-branching-model.md` |
| **coin-api / coin-ui / coin-executor** | без обязательных изменений кода (SoT уже там) |
| **Local pilot** | seed и E2E должны остаться green после переноса fixtures |

Связанные ADR: [gp-branching-model](../../docs/adr/gp-branching-model.md), [control-plane-v2](../../docs/adr/control-plane-v2.md).
