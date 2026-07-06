## Why

Текущий редактор Build Stack больше не соответствует жёсткому Golden Path: он смешивает YAML и JSON preview, показывает дублирующие представления, не даёт полноценно настраивать deliverables, параметры и цели сборки, а manifest остаётся слишком грубым для реального build/publish runtime.

После перехода к hard GP Build Stack должен стать явной платформенной моделью: enabling team настраивает outputs, параметры, цели, Containerfile artifacts и execution contract в UI, а product repo только выбирает GP version.

## What Changes

- **BREAKING**: заменить текущий `content.yaml` v2 editor на Build Stack vNext editor без режима «сырой YAML как основной UX».
- **BREAKING**: manifest `build.engine` перестаёт быть единственным engine на весь stack; engine выбирается на уровне build target / deliverable step.
- **BREAKING**: `capabilities.deliverables: []` заменяется структурированной моделью deliverables с именем, типом, build target, publish policy, artifact inputs и Containerfile source.
- **BREAKING**: `pipeline.stages` получают связь с параметрами, deliverables и build targets; stage больше не является только typed id/name списком.
- Добавить раздел Build Stack parameters: типизированные параметры с default/required/allowed values, которые executor сможет использовать в stages, targets и templates.
- Добавить model для Containerfile artifacts: каждый managed Containerfile хранится в gp-content package как named artifact и материализуется в manifest content refs.
- Убрать дублирующий preview YAML/JSON: UI показывает canonical model form + resolved manifest preview; YAML export/debug не входит в P0.
- Сформировать новый resolved manifest contract для executor: `parameters`, `deliverables`, `build.targets`, `pipeline.stages`, `artifacts.containerfiles`, `destinations`.
- Сохранить правило: product config не задаёт build/publish logic, credentials не попадают в manifest, `coin-lib` остаётся glue-only.

## Capabilities

### New Capabilities

- `build-stack-vnext`: новая платформенная модель Build Stack: параметры, targets, deliverables, artifacts и связи между ними.
- `manifest-build-contract-vnext`: новый resolved manifest contract для executor после структурирования Build Stack.
- `build-stack-ui-vnext`: UX создания/редактирования Build Stack без YAML/JSON дублирования и с полноценной настройкой deliverables/parameters/targets.

### Modified Capabilities

- `platform-build-stacks`: заменить editor requirements с `content.yaml` v2 card editor на Build Stack vNext editor.
- `gp-content-preview`: preview должен возвращать resolved manifest preview из canonical model, без physical destination и без дублирования YAML/JSON как равных источников.
- `build-engine`: engine dispatch должен работать на уровне target/deliverable step, а не только top-level `build.engine`.
- `gp-composition-two-slot`: `gp-content` pin остаётся одним slot, но материализуемый build stack contract внутри этого pin становится vNext.

## Impact

- `coin-ui`: полная переделка Build Stack create/edit/detail экранов, preview panel, forms для parameters/targets/deliverables/Containerfiles.
- `coin-api`: schema/model validation для Build Stack vNext, package content_ref, preview API, OpenAPI, manifest builder.
- `coin-gp-content`: новая schema vNext и seed content для pilot stack.
- `coin-executor`: manifest structs, target engine dispatch, parameter resolution, deliverables execution, Containerfile materialization.
- `coin-lib`: без бизнес-логики; только передаёт manifest/config/credentials в executor stages.
- Docs/ADR: зафиксировать Build Stack vNext как новый permanent contract после hard GP.

## Non-goals

- Не возвращаем build/publish logic в product repo.
- Не делаем Jenkins config source of truth для parameters, destinations или routing.
- Не вводим отдельный destination service/catalog в этом change.
- Не проектируем все языковые экосистемы package publish; P0 остаётся image/liquibase-image/artifact.
- Не реализуем arbitrary shell commands от продуктовой команды.
- Не меняем трёхслотовую GP composition: `agent`, `gp-content`, `branching-model`.
