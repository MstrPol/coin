# ADR: Build Stack vNext contract

**Статус:** superseded ([pipeline-inline-build-stack.md](./pipeline-inline-build-stack.md))  
**Дата:** 2026-07-02  
**Связанный change:** `rework-build-stack-contract`

## Контекст

После перехода к hard Golden Path продуктовый репозиторий больше не должен определять build/publish logic. Product config хранит GP pin, project identity и Jenkins glue, а Build Stack внутри `gp-content` задаёт параметры, цели сборки, deliverables, Containerfile artifacts и pipeline contract.

Текущий editor и manifest contract были слишком близки к `content.yaml`: UI показывал форму, YAML и JSON preview как почти равные представления, deliverables были плоским списком capabilities, `build.engine` выбирался один раз на stack, а Containerfile artifact был один на все случаи.

## Решение

### Canonical model

Build Stack vNext хранится как structured model в `gp-content` package. UI редактирует модель через формы и не делает raw YAML основным или альтернативным source of truth.

В P0 нет YAML export/debug режима. Основной preview — resolved manifest и validation issues.

### Parameters

Build Stack задаёт типизированные non-secret parameters:

- `string`
- `boolean`
- `number`
- `enum`

Parameters могут использоваться в target config, stage conditions и template fields. Secret values и Jenkins credential IDs не являются parameters и остаются в product/Jenkins glue.

### Build targets

`build.engine` больше не является единственным dispatch source. Build Stack задаёт named targets, и каждый target выбирает engine самостоятельно:

```yaml
build:
  targets:
    - id: app-image
      engine: buildkit
      containerfile: app
      target: runtime
    - id: migrations-image
      engine: dockerfile
      dockerfile: db/Dockerfile
      target: runtime
```

### Deliverables

Deliverables становятся named outputs с типом, target reference и publish metadata:

```yaml
deliverables:
  - id: app
    type: image
    targetId: app-image
  - id: liquibase
    type: image
    targetId: migrations-image
    image:
      repositorySuffix: "-liquibase"
```

В P0 разрешены несколько named deliverables одного type. Ограничение «один output каждого type» не закрепляется в UI и manifest contract.

### Containerfile artifacts

Managed Containerfiles хранятся как named artifacts в `gp-content` package и материализуются в manifest immutable content refs:

```yaml
artifacts:
  containerfiles:
    - id: app
      path: dockerfiles/app.Containerfile
    - id: liquibase
      path: dockerfiles/liquibase.Containerfile
```

Executor материализует только те Containerfile artifacts, которые нужны target.

### Pipeline stages

Stages остаются typed executor stages, но получают steps со ссылками на platform actions, targets и deliverables. Arbitrary shell scripts от product repo или Jenkins Shared Library не допускаются.

```yaml
pipeline:
  stages:
    - id: build
      steps:
        - action: build-deliverable
          deliverableId: app
```

### Manifest

Отдельный `manifestVersion: 2` и `build.schemaVersion` не вводятся. Для local pilot применяется hard cut текущей формы manifest: resolved manifest получает новые sections `parameters`, `build.targets`, `deliverables`, `artifacts.containerfiles`, `pipeline.stages` и продолжает содержать integrity metadata.

Manifest должен быть самодостаточным для build path через Nexus fallback и не должен содержать credentials, secret values или Jenkins credential IDs.

## Последствия

- `coin-api` валидирует Build Stack vNext как structured model и материализует deterministic manifest.
- `coin-ui` получает model-first editor без YAML/JSON дублирования.
- `coin-executor` исполняет stage steps из manifest и dispatch по target-level engine.
- `coin-lib` остаётся Jenkins glue only: resolve/fallback, pod, credentials binding, вызов executor stages.
- Product repo не получает build/publish commands, deliverables или parameters как source of truth.

## Отклонённые альтернативы

| Альтернатива | Почему отклонена |
|--------------|------------------|
| Оставить YAML editor + JSON preview | Сохраняет дубли и непонятный source of truth |
| YAML edit mode с round-trip | Усложняет validation и возвращает второй контракт рядом с формой |
| Один top-level `build.engine` | Не позволяет настраивать engine для каждой цели |
| Один deliverable каждого type | Слишком тесно для hard GP и заставляет менять контракт при первом multi-output stack |
| `manifestVersion: 2` | Для local pilot проще hard cut незашипленного manifest shape |
