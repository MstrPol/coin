# Build stacks (pipeline-inline v3)

Редактирование build stack — через **Component Studio** (`/platform/build-stacks/.../edit`) как structured model. Raw YAML не является primary UX.

ADR: [pipeline-inline-build-stack](../adr/pipeline-inline-build-stack.md).

## Основные секции

| Секция | Назначение |
|--------|------------|
| Parameters | Типизированные non-secret параметры: `string`, `boolean`, `number`, `enum` |
| Pipeline stages | Stages со steps `run`, `build`, `publish`; buildkit steps содержат inline `containerfile.body` |
| Preview | Resolved manifest preview + validation issues |

**Нет** отдельных секций `build.targets`, `deliverables`, `artifacts.containerfiles` — всё inline в pipeline steps.

Buildpack **не поддерживается** (hard cut).

## Минимальная модель (`schemaVersion: 3`)

Пример Build Stack для `go-app`:

```yaml
schemaVersion: 3
name: go-app
kind: gp-content
validateSchema: schemas/config.v2.schema.json

parameters:
  - name: GO_VERSION
    type: string
    default: "1.22"
    required: true

pipeline:
  stages:
    - id: validate
      name: Validate
      steps:
        - action: run
          run:
            engine: buildkit
            target: validate
            output: validate
            containerfile:
              body: |
                FROM golang:1.22-bookworm AS base
                ...
    - id: build
      name: Build
      steps:
        - action: build
          build:
            id: app
            type: image
            engine: buildkit
            target: runtime
            containerfile:
              body: |
                FROM golang:1.22-bookworm AS base
                ...
        - action: publish
          publish:
            buildStepId: app
```

### Step actions

| action | Назначение |
|--------|------------|
| `run` | validate / test / промежуточный target |
| `build` | materialize output (`build.id`, `type`, engine config) |
| `publish` | publish по `publish.buildStepId` → `build.id` |

Machine ids (`stage.id`, `build.id`) — short hash **5–6 символов** `^[a-z0-9]{5,6}$`; UI генерирует при создании. Человекочитаемые имена — в `stage.name`.

### Containerfile inline

Buildkit `run` / `build` steps содержат `containerfile.body` в author model. Resolved manifest materializer добавляет на тот же step `containerfile.contentRef` + `digest` (без top-level catalog).

### BYO Dockerfile

Для product-owned Dockerfile (`go-app-docker`):

```yaml
- action: build
  build:
    id: app
    type: image
    engine: dockerfile
    dockerfile:
      path: Dockerfile
      target: runtime
```

Продуктовый `config.yaml` **не** задаёт путь к Dockerfile — только GP pin и project identity.

`content.yaml` не содержит registry/cache/artifact repository URLs. Physical destinations задаются полями GP version и материализуются в manifest `destinations`.

## Validate и preview

| API | Назначение |
|-----|------------|
| `POST /v1/admin/components/gp-content/.../validate` | Draft package (content.yaml + validateSchema artifact) |
| `POST /v1/admin/gp-content/preview` | Manifest subset preview |

Studio вызывает preview при редактировании pipeline steps.

## Publish flow

1. Author в Studio → pipeline-inline v3 model (containerfile inline в buildkit steps).
2. Validate → Register → Promote component version.
3. Pin `gp-content` в GP composition draft.
4. GP promote → resolve manifest.

Local bootstrap: `cd docker && make seed-jenkins-lib`.

## E2E

```bash
cd docker
make e2e-build-engines   # demo-go-app (buildkit) + demo-go-app-docker (BYO)
```

См. также [agent-build-model.md](../agent-build-model.md), ADR [build-engine-contract](../adr/build-engine-contract.md).
