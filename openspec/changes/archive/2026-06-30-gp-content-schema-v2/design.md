## Context

**Branching (done):** `model.yaml` v2 → cards editor → preview API → executor SoT.

**gp-content (today):**

```yaml
# legacy reference stacks
controls: {...}           # dead — не в runtime
pipeline.stages[].when: tag   # superseded branching + params.publish
build.engine: buildkit | buildpack | dockerfile
```

- UI: partial form + `parseGpContentYaml` line parser; presets только buildkit.
- `dockerfile` engine = managed Containerfile (как buildkit-lite) — **не** BYO.
- buildpack = отдельный E2E path, тяжёлый agent bootstrap.

**Platform lead decisions (2026-06 explore):**

| # | Решение |
|---|---------|
| Q1 | BYO Dockerfile — **отдельный GP profile**; в проекте только `goldenPath` + `version` |
| Q2 | buildpack — **hard cut** |
| Q3 | go-app-df + old dockerfile engine — **удалить** |
| Q4 | `artifact` deliverable — **только buildkit**; BYO = image only |

## Goals / Non-Goals

**Goals:**

- `content.yaml` v2 как SoT для build stack editor (как branching v2).
- Два engine в schema, UI, validate, executor, E2E.
- Preview API = coin-api + manifest builder subset (не дублировать в TS).
- Hard cut v1 content.yaml и buildpack.

**Non-goals:**

- Project config build overrides.
- Три engine / buildpack fallback.
- Corp buildkitd migration (отдельно).

## Decisions

### D1. schemaVersion 2 hard cut

Как branching v2: `schemaVersion: 2` обязателен; без него — reject.

### D2. YAML shape (v2)

```yaml
schemaVersion: 2
name: go-app
kind: gp-content

capabilities:
  deliverables:
    - image
    - artifact          # только buildkit GP

build:
  engine: buildkit | dockerfile
  buildkit:
    targets:
      validate: validate
      test: test
      image: runtime
      artifact: artifact
    cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit"
  dockerfile:           # BYO GP only
    path: Dockerfile    # relative to workspace checkout
    imageTarget: runtime
    testTarget: test
    cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:dockerfile"

pipeline:
  stages:
    - id: validate
      name: Validate
    - id: test
      name: Test
    - id: build
      name: Build
    - id: publish
      name: Publish

artifacts:
  validateSchema: schemas/config.v2.schema.json
  containerfile: dockerfiles/Containerfile   # required buildkit only
```

- **Нет** `controls`.
- **Нет** `pipeline.stages[].when`.
- `artifacts.containerfile` — только `buildkit` (managed Containerfile в gp-content package).
- `build.buildkit.dockerfile` path **не** дублировать — runtime всегда `.coin/Containerfile` после materialize.

### D3. Два engine

| Engine | GP | Containerfile source | Deliverables | materialize |
|--------|-----|----------------------|--------------|-------------|
| `buildkit` | `go-app` | gp-content artifact → `.coin/Containerfile` | image + artifact | да |
| `dockerfile` | `go-app-docker` * | product repo `build.dockerfile.path` | image only | **нет** |

\* имя профиля — в tasks; не `go-app-df`.

Продукт выбирает модель pin'ом `coin.goldenPath`, без полей в `config.yaml`.

### D4. Buildpack hard cut

Удалить: `internal/build/buildpack.go` usage, `COIN_BUILD_ENGINE=buildpack` bootstrap, `go-app-bp`, demo job, agent `pack`/tar.

### D5. Preview API

`POST /v1/admin/gp-content/preview`

Request: draft `content.yaml` body (или manifest subset) + optional `engine` override for dry-run.

Response: resolved fragment `{ build, pipeline, capabilities }`, validation issues, warnings (e.g. BYO without testTarget).

Executor **не** обязателен для v1 preview — manifest builder + validate rules достаточно; опционально dry-run warnings из shared validate package.

**Альтернатива:** full resolve mock — отклонено для scope; subset как branching preview.

### D6. UI editor (branching parity)

Cards:

1. **Engine** — `buildkit` | `dockerfile` (один active block)
2. **Build policy** — targets (buildkit) или path/targets (BYO)
3. **Capabilities** — deliverables checklist (enforce engine rules)
4. **Pipeline stages** — ordered list
5. **Artifacts** — schema key; containerfile key (buildkit only) + Containerfile tab
6. **Preview** — manifest snippet + warnings

Presets: `go-app` (buildkit), `go-app-docker` (BYO).

### D7. Publish gate (docs only)

Не в gp-content schema. См. branching + `params.publish` ([coin-ci-runtime](../../docs/adr/coin-ci-runtime.md)).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| BYO Dockerfile нарушает corp standards | validate stage + future Dockerfile policy lint |
| Hard cut ломает go-app-bp pins | нет prod fleet; seed обновить |
| Два GP на Go (`go-app` vs `go-app-docker`) | docs how-to «когда какой pin» |
| Preview без executor | validate + manifest builder; expand later |

## Migration Plan

1. Schema JSON + validate-package v2
2. Executor + agent buildpack removal; BYO dockerfile path
3. coin-gp-content stacks migrate/delete
4. Preview API
5. UI editor rewrite
6. E2E 2 engines; docs/ADR amend
7. Archive change

## Open Questions

| # | Вопрос | Статус |
|---|--------|--------|
| Q1 | Имя BYO GP profile | ⏳ default `go-app-docker` в tasks |
| Q2 | Preview через executor или builder only | ✅ builder + validate (D5) |
