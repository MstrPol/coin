# Build stacks (gp-content v2)

Редактирование build stack — через **Component Studio** (`/platform/build-stacks/.../edit`) или YAML `content.yaml` в `coin-gp-content/stacks/<name>/`.

**Schema:** `coin-gp-content/schemas/gp-content.schema.json` (`schemaVersion: 2`).

## Два build engine

| Engine | GP profile | Containerfile | Deliverables |
|--------|------------|---------------|--------------|
| `buildkit` | `go-app` | Managed в gp-content package (`dockerfiles/Containerfile`) | `image`, `artifact` |
| `dockerfile` (BYO) | `go-app-docker` | Dockerfile **в репозитории продукта** | `image` only |

Buildpack и managed `dockerfile` engine (старый `go-app-df`) **не поддерживаются** (hard cut).

## content.yaml v2 (минимум)

### buildkit (`go-app`)

```yaml
schemaVersion: 2
name: go-app
kind: gp-content
capabilities:
  deliverables: [image, artifact]
build:
  engine: buildkit
  buildkit:
    targets:
      validate: validate
      test: test
      image: runtime
      artifact: artifact
    cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit"
pipeline:
  stages:
    - { id: validate, name: Validate }
    - { id: test, name: Test }
    - { id: build, name: Build }
    - { id: publish, name: Publish }
artifacts:
  validateSchema: schemas/config.v2.schema.json
  containerfile: dockerfiles/Containerfile
```

### BYO dockerfile (`go-app-docker`)

```yaml
schemaVersion: 2
name: go-app-docker
kind: gp-content
capabilities:
  deliverables: [image]
build:
  engine: dockerfile
  dockerfile:
    path: Dockerfile
    imageTarget: runtime
    testTarget: test
    cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:dockerfile"
pipeline:
  stages: [...]
artifacts:
  validateSchema: schemas/config.v2.schema.json
```

Продуктовый `config.yaml` **не** задаёт путь к Dockerfile — только `coin.goldenPath: go-app-docker`.

## Validate и preview

| API | Назначение |
|-----|------------|
| `POST /v1/admin/components/gp-content/.../validate` | Draft package (content.yaml + artifacts) |
| `POST /v1/admin/gp-content/preview` | Manifest subset preview (как branching preview) |

Studio вызывает preview при редактировании карточек.

## Publish flow

1. Author в Studio → draft `content.yaml` + `dockerfiles/Containerfile` (buildkit only).
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
