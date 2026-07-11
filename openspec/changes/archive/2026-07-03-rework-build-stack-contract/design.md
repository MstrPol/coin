## Context

`gp-manifest-publish-routing` заархивирован как промежуточный change без синхронизации specs: он решил часть вопросов про destinations и thin product config, но показал, что Build Stack contract требует полной переработки.

Текущий Build Stack editor всё ещё мыслит `content.yaml` как основной артефакт. UI показывает форму, YAML и JSON preview как почти равные представления, из-за чего появляются дубликаты и непонятно, что является source of truth. Deliverables представлены как простой список capabilities, Containerfile один на stack, parameters отсутствуют, а `build.engine` выбирается один раз на весь stack.

Hard Golden Path требует другой модели: product repo не пишет build/publish logic, enabling team описывает Build Stack в платформе, `coin-api` материализует deterministic manifest, а `coin-executor` исполняет его без live DB.

## Goals / Non-Goals

**Goals:**

- Сделать Build Stack vNext canonical model source of truth вместо YAML-first редактора.
- Дать UI полноценные разделы: parameters, targets, deliverables, Containerfile artifacts, pipeline stages, manifest preview.
- Поддержать target-level engine selection: разные цели могут использовать разные engines и свои build inputs.
- Связать deliverables с targets, artifacts, publish policy и outputs.
- Материализовать все runtime-needed данные в manifest для Nexus fallback.
- Убрать дублирующий YAML/JSON preview: YAML только export/debug, manifest preview — результат materialization.
- Сохранить `coin-lib` glue-only и отсутствие credentials в manifest.

**Non-Goals:**

- Не переносить build/publish commands в product repo.
- Не делать Jenkins source of truth для parameters, destinations или stages.
- Не вводить отдельный destination catalog/service.
- Не проектировать fleet rollout до corp gate.
- Не поддерживать произвольные shell commands от продуктовых команд.
- Не расширять P0 package ecosystems за пределы image/liquibase-image/artifact.

## Decisions

### Решение 1: canonical model вместо YAML-first editor

Build Stack vNext хранится как structured model в gp-content package. UI редактирует эту модель через формы. `content.yaml` не является главным UX и в P0 не показывается как export/debug режим.

Альтернативы:

- **A: оставить YAML editor + preview** — дешевле, но сохраняет текущую путаницу и дубли.
- **B: canonical form model + optional export** — лучше текущего состояния, но оставляет второй способ чтения контракта.
- **C: canonical form model без YAML export/debug в P0** — жёстче, зато убирает дубли и снижает риск расхождения представлений.

Выбор: C.

### Решение 2: targets — центральная единица build execution

`build.engine` top-level заменяется на `build.targets[]`. Каждый target имеет `id`, `engine`, engine-specific config, inputs, optional Containerfile artifact и outputs. Pipeline stages и deliverables ссылаются на targets по `targetId`.

Пример модели:

```yaml
build:
  targets:
    - id: app-image
      engine: buildkit
      containerfile: app
      target: runtime
    - id: app-artifact
      engine: buildkit
      containerfile: app
      target: artifact
    - id: migrations-image
      engine: dockerfile
      dockerfile: db/Dockerfile
      target: runtime
```

### Решение 3: deliverables становятся структурированными outputs

Deliverables больше не являются `capabilities.deliverables: [image, artifact]`. Они становятся named entities:

```yaml
deliverables:
  - id: app
    type: image
    targetId: app-image
    image:
      repositorySuffix: ""
  - id: liquibase
    type: liquibase-image
    targetId: migrations-image
    image:
      repositorySuffix: "-liquibase"
  - id: app-zip
    type: artifact
    targetId: app-artifact
    artifact:
      format: zip
      paths:
        - /out/app
```

В P0 разрешены несколько named deliverables одного type. Это нужно, чтобы Build Stack не зашивал ограничение «один image / один artifact» в UI и manifest contract.

### Решение 4: parameters — типизированный runtime contract

Build Stack задаёт параметры типов `string`, `boolean`, `number` и `enum`, которые попадают в manifest:

```yaml
parameters:
  - name: GO_VERSION
    type: string
    default: "1.22"
    required: true
  - name: RUN_TESTS
    type: boolean
    default: true
```

Parameters можно использовать в target config, stage conditions и template fields. Secrets не являются parameters; credential IDs остаются в product/Jenkins glue.

### Решение 5: Containerfile artifacts — named artifacts в package и manifest

У stack может быть несколько managed Containerfile artifacts:

```yaml
artifacts:
  containerfiles:
    - id: app
      path: dockerfiles/app.Containerfile
    - id: liquibase
      path: dockerfiles/liquibase.Containerfile
```

Manifest содержит immutable content refs для каждого managed Containerfile. Executor материализует нужный artifact по ссылке target/containerfile.

### Решение 6: pipeline stages связываются с targets/deliverables/parameters

Stages остаются typed, но получают `steps` как ссылки на платформенные actions, а не arbitrary shell:

```yaml
pipeline:
  stages:
    - id: test
      steps:
        - action: run-target
          targetId: app-test
    - id: build
      steps:
        - action: build-deliverable
          deliverableId: app
        - action: build-deliverable
          deliverableId: app-zip
```

Это сохраняет запрет на GP shell scripts и не переносит business logic в Jenkins.

### Решение 7: preview показывает только materialization result

Preview panel показывает:

- validation issues;
- warnings;
- resolved manifest preview;
- optional package artifact list.

UI не показывает одновременно raw YAML и raw JSON как два редактируемых источника. В P0 YAML export/debug не нужен; основной preview — resolved manifest и validation issues.

## Risks / Trade-offs

- **Сложность UI растёт** → разделить editor на cards: Parameters, Targets, Deliverables, Containerfiles, Pipeline, Preview.
- **Executor contract резко меняется** → ввести manifest vNext внутри hard cut без backward shim для незашипленных branch changes.
- **Parameters могут смешаться с secrets** → явно запретить secret values и credential IDs в parameters; credentials только Jenkins glue.
- **Target-level engines усложняют validation** → validation должна проверять ссылки `stage.step.targetId`, `deliverable.targetId`, `containerfile`.
- **Несколько Containerfile artifacts усложняют package** → хранить их как named artifact bodies в gp-content package и content_ref.

## Migration Plan

### Migration note

`gp-manifest-publish-routing` заархивирован без sync specs как промежуточный change. Его implementation changes в рабочем дереве могут быть переиспользованы только если они не противоречат Build Stack vNext contract; старые ограничения вроде flat deliverables, top-level `build.engine` и «один output каждого type» не считаются целевым контрактом.

1. Зафиксировать ADR `build-stack-vnext-contract`.
2. Добавить Build Stack vNext schema и model types в `coin-api`.
3. Переделать `gp-content` package content_ref под structured model vNext.
4. Переделать Build Stack editor в `coin-ui`.
5. Добавить preview materializer из vNext model в manifest preview.
6. Обновить manifest schema/OpenAPI.
7. Переделать `coin-executor` manifest structs и execution dispatch.
8. Пересеять pilot Build Stack / GP release.
9. Прогнать `samples/demo-go-app` E2E.

Rollback для local pilot: вернуться к архивному состоянию branch до применения change. Для corp/fleet rollout этот change не применяется до corp gate.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Нужен ли в P0 UI режим YAML export/debug? | ✅ decided | A: только read-only export; B: edit YAML с round-trip; C: убрать полностью | C: убрать полностью в P0 |
| Q2 | Какие parameter types входят в P0? | ✅ decided | A: string/boolean/number/enum; B: только string/boolean; C: плюс arrays/maps | A: string/boolean/number/enum |
| Q3 | Как назвать manifest vNext: `manifestVersion: 2` или `build.schemaVersion` внутри v1? | ✅ decided | A: manifestVersion 2; B: build.schemaVersion; C: hard cut v1 shape | C: hard cut текущей формы manifest |
| Q4 | Нужны ли несколько deliverables одного type уже в P0? | ✅ decided | A: оставить максимум один type; B: разрешить несколько named outputs сразу | B: разрешить несколько named outputs одного type |
| Q5 | Catalog-first UI vs pipeline inline config? | ✅ decided | A: pipeline-first + hidden catalogs; B: wizard; C: inline config в steps (contract change) | C → новый change `pipeline-inline-steps` |

## Supersession

Implementation P0 завершён (секции 1–5, catalog editor). Pilot/E2E (tasks 6.2–6.3, 6.6, 6.8) и catalog-first UX **не доводятся** — superseded change **`pipeline-inline-steps`**: pipeline stages с inline build/publish config.
