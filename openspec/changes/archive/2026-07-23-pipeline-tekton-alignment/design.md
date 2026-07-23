## Context

После `gp-release-embedded-pipeline` pipeline-inline **v3** живёт на GP release (`stages`, inline Containerfile). Enabling team согласовала Tekton mental model на Jenkins runtime. Параллельно нужна **быстрая петля отладки executor**: править resolved-манифест в репо без API/Nexus.

Связанные ADR: [pipeline-inline-build-stack.md](../../docs/adr/pipeline-inline-build-stack.md) (v3, supersede), [control-plane-v2.md](../../docs/adr/control-plane-v2.md), [jenkins-lib-http-nexus.md](../../docs/adr/jenkins-lib-http-nexus.md).

## Goals / Non-Goals

**Goals:**

- `schemaVersion: 4` — `pipeline.tasks[]`, `runAfter`, step kinds (`coin`, `containerfile`, `sh`), `containerfiles[]` catalog.
- Product config: `coin.resolve: file | remote`, `coin.manifestFile` (default `.coin/manifest.local.yaml`).
- coin-lib: file resolve + soft warn; linear task DAG → Jenkins stages; `--task`.
- coin-executor: v4 dispatch.
- Fixture-driven acceptance (sample/demo) **без** live control plane / UI — file resolve only.

**Out of scope (отдельный change [`pipeline-v4-control-plane`](../pipeline-v4-control-plane/)):**

- coin-api validate/preview/materialize v4 + v3 migration.
- coin-ui task/catalog editors.
- Remote resolve E2E через seed.

**Non-Goals:**

- Tekton Controller, CRDs, Triggers.
- Отдельный pin для containerfile catalog.
- Смена 2-pin composition.
- Hard block `resolve: file` в product repos.
- Corp fleet migration.

## Decisions

### D1: schemaVersion 4 — `pipeline.tasks` вместо `pipeline.stages`

Author / resolved model:

```yaml
schemaVersion: 4
branching: {...}
parameters: [...]
validateSchema: ...
# agent/runtime, destinations — peer-секции resolved manifest
containerfiles:
  - id: app
    kind: managed
    path: .coin/containerfiles/app
    # body / contentRef — не в resolved/file fixture; файл на диске по path
  - id: liquibase
    kind: project
    path: docker/liquibase.Containerfile
pipeline:
  tasks:
    - id: validate
      name: Validate
      steps:
        - kind: coin
          action: validate
    - id: test
      name: Test
      runAfter: [validate]
      steps:
        - kind: containerfile
          ref: app
          run:
            target: test   # Containerfile AS test; не agent cmd
    - id: build-app
      runAfter: [test]
      steps:
        - kind: coin
          action: build
          build:
            type: image
            containerfileRef: app   # только catalog id
    - id: publish-app
      runAfter: [build-app]
      when: tag
      steps:
        - kind: coin
          action: publish
          publish:
            buildTaskId: build-app  # id build-task → .coin/outputs.json
```

Runtime остаётся Jenkins. Fixture Phase A — **уже resolved** документ (те же peer-секции + `pipeline.tasks`), не сырой GP draft без pins.

### D2: Top-level `containerfiles[]` catalog (variant B)

Секция `containerfiles` — peer к `branching` / `parameters` / `agent` (runtime) / `destinations`, не внутри pipeline step.

Каждая запись: **`id`**, **`kind`**, **`path`** (workspace-relative). Build/test steps ссылаются **только на `id`**.

| `kind` | Смысл | Runtime materialize |
|--------|-------|---------------------|
| `managed` | Контент с платформы | Fetch (`contentRef`) **или** authoring `body` → **записать в `path`** |
| `project` | BYO файл в product repo | Использовать файл по `path` as-is |

**Workspace layout (обязательный контракт path):**

```
.coin/
  containerfiles/
    app              # файл на диске — SoT содержимого для runtime/fixture
    liquibase        # …
  manifest.local.yaml   # catalog: id + kind + path (без inline body)
  manifest.json         # runtime copy после resolve
```

Правила:

- `path` обязателен для обоих kind; default для managed: `.coin/containerfiles/<id>`.
- Дубликат `id` или два entry на один `path` → validate reject.
- Step MUST NOT содержать inline Containerfile body.
- Step kinds `coin` build / `containerfile` используют `containerfileRef` / `ref` = catalog `id`.

**Где живёт содержимое (не путать слои):**

| Слой | `body` в YAML? | Файлы в `.coin/containerfiles/` |
|------|----------------|----------------------------------|
| GP authoring (Phase B UI/API) | да, в draft body / storage | появляются после materialize |
| Resolved remote manifest | нет — `contentRef`+`digest` + `path` | executor пишет по `path` при run |
| File-resolve fixture (`resolve: file`) | **нет** — только `id`/`kind`/`path` | **да**, файлы в git/workspace рядом с fixture |

File fixture и resolved view **SHALL NOT** встраивать полный Containerfile в `containerfiles[].body`. Содержимое — отдельные файлы по `path` (обычно `.coin/containerfiles/<id>`).

**Отклонено:** `kind: coin` как имя платформенного контента — оставляем `managed` | `project` (variant B).
**Отклонено:** inline `body` в `manifest.local.yaml` / resolved fixture — только path + disk file.

### D3: Step kinds

| kind | Phase A | Описание |
|------|---------|----------|
| `coin` | да | `validate`, `test`, `build`, `publish` |
| `containerfile` | да | invoke по catalog ref |
| `sh` | pilot allowlist | только allowlisted scripts |

### D4: Semantic task ids

`task.id` = `^[a-z][a-z0-9-]{1,31}$`.  
`coin-executor run --task <id>` (`--stage` deprecated alias).

### D5: `runAfter` DAG

- Phase A: линейный порядок + явный `runAfter`; coin-lib — последовательные stages.
- Later: Jenkins `parallel` по frontier.

### D6: Manifest shape self-sufficiency

Resolved manifest (file или remote) MUST быть достаточен для executor без live PG. Managed containerfiles — из catalog/`contentRef` или файла по `path`.

**Не включать в v4:** `capabilities.deliverables`, top-level `deliverables`, `build.targets`, legacy `artifacts.containerfiles` — build shape задаёт `pipeline.tasks`.

### D7: UI layout

Вынесено в [`pipeline-v4-control-plane`](../pipeline-v4-control-plane/) (Composition → Pipeline → Containerfiles → Parameters).

### D8: Migration v3→v4

Вынесено в [`pipeline-v4-control-plane`](../pipeline-v4-control-plane/) (read adapter + migrate on save + seed).

### D9: File resolve

Product `.coin/config.yaml`:

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
  resolve: file
  # manifestFile: .coin/manifest.local.yaml  # default
```

| Поле | Default | Смысл |
|------|---------|--------|
| `coin.resolve` | `remote` (или отсутствует) | `remote` → API→Nexus; `file` → локальный файл |
| `coin.manifestFile` | `.coin/manifest.local.yaml` | путь от корня проекта; только при `file` |

Правила:

- `goldenPath` + `version` **обязательны** и при `file` (identity / report / будущий promote).
- Resolve path: прочитать YAML/JSON fixture → вернуть map → `coinMaterializeDotCoin` пишет `.coin/manifest.json` (gitignore) как сегодня.
- Soft warn в логе Jenkins при `resolve: file` (не hard fail).
- Fixture **не** путать с runtime `.coin/manifest.json`.
- Формат fixture = resolved shape (то, что ест executor), включая `runtime` / `branching` / `destinations` / `pipeline.tasks` / `containerfiles`.

```
coinResolveManifest(cfg)
├─ resolve == file → read manifestFile → soft warn → return map
└─ resolve == remote (default) → API → Nexus (без изменений)
```

**Альтернатива A:** `goldenPath: local`. Отклонено — смешивает GP identity и источник bytes.  
**Альтернатива B:** только env. Отклонено — не видно в repo.  
**Выбор:** отдельные поля `resolve` / `manifestFile`.

### D10: Delivery order — executor-first

```
pipeline-tekton-alignment          pipeline-v4-control-plane
─────────────────────────          ─────────────────────────
JSON Schema v4 artifact            coin-api validate/storage
config v2 + file resolve           materialize / preview
executor v4 + --task               v3→v4 migration
coin-lib tasks + file              UI editors
fixture + offline sample build     seed PG + remote E2E
ADR mapping
```

API/UI подстраиваются под стабильный fixture contract (отдельный change).

### D11: Build / publish I/O contract

Executor — единственный исполнитель actions. Authoring **не** задаёт полный `imageRef` в step (кроме редкого debug override вне GP contract).

**Build (`kind: coin`, `action: build`):**

| Input | Обязателен | Источник |
|-------|------------|----------|
| `containerfileRef` | да | `containerfiles[].id` |
| `type` | да (`image`, …) | step |
| `destinationRef` | да | `destinations[].id` |
| `cache` | нет | override `destinations[].buildCacheEnabled` |
| `imageRef` | нет | **computed** executor |

Computed image ref:

```
{destination.imageRegistryPrefix}/{project.groupId}/{project.artifactId}/{project.name}:{tag}
```

`tag` — branching versioning / `COIN_IMAGE_TAG` / fallback GP version.

После успешного build executor **мержит** запись в единый `.coin/outputs.json`:

```json
[
  { "name": "build-app", "type": "image", "ref": "registry/.../my-app:tag" }
]
```

`name` записи = `pipeline.tasks[].id` build-task (ключ для publish).

**Publish (`kind: coin`, `action: publish`):**

| Input | Обязателен | Смысл |
|-------|------------|-------|
| `buildTaskId` | да | id build-task → entry в `.coin/outputs.json` |
| `destinationRefs` | да (non-empty) | multi-push в catalog entries с `push: true` |

- Publish **не** принимает `containerfileRef` и **не** требует литеральный `imageRef` в YAML.
- Для каждого destination: при необходимости retag под его `imageRegistryPrefix`, затем push.
- Gates: `manifest.branching` publish policy + Jenkins `params.publish`.
- N build-tasks и N publish-tasks допустимы.

### D13: Destinations catalog (auth + build/publish refs)

Две цели top-level `destinations`:

1. **Pre-auth** — login во все registry с `pull: true` или `push: true` до test/build (coin-lib).
2. **Naming / publish** — build и publish ссылаются на entry по `id`.

```yaml
destinations:
  - id: nexus-docker
    host: localhost:8082
    imageRegistryPrefix: localhost:8082/coin-docker
    pull: true
    push: true
    buildCacheEnabled: false
    artifactRepositoryBase: http://nexus:8081/repository/maven-releases   # optional

# build
build:
  type: image
  containerfileRef: app
  destinationRef: nexus-docker
  cache: false   # optional override of destination.buildCacheEnabled

# publish
publish:
  buildTaskId: build-app
  destinationRefs: [nexus-docker]   # multi-push; retag if prefix ≠ build ref
```

| Поле entry | Смысл |
|------------|--------|
| `id` | semantic id; refs из steps |
| `host` | registry host для docker login (иначе из prefix) |
| `imageRegistryPrefix` | prefix для computed image/cache ref |
| `pull` / `push` | участие в bootstrap auth; publish требует `push: true` |
| `buildCacheEnabled` | default cache для build с этим destinationRef |
| `artifactRepositoryBase` | optional, для artifact publish later |

Legacy flat `destinations: { imageRegistryPrefix, … }` — только pre-v4 remote manifests.

### D12: Test-in-container + report export

**Инвариант:** language toolchains **нет** на `coin-agent`. Юнит-тесты выполняются только через Containerfile target (как сегодня go-app `AS test` + `RUN go test`).

Default Phase A path:

1. Catalog entry (`managed`|`project`) → materialize/use `path`.
2. Executor: `podman build -f <path> --target <testTarget>` (имя target из step, default `test`).
3. Step authoring — `kind: containerfile` + `ref` + `run.target`, **или** сахар `kind: coin` / `action: test` с `containerfileRef` + `target` (эквивалент).
4. **Не** default: `inputs.cmd` / shell на агенте; Testcontainers + docker socket — **вне Phase A**.

**Отчёты как артефакт Jenkins:**

Файлы из `RUN` внутри build не попадают в workspace сами. Контракт:

| Роль | Обязанность |
|------|-------------|
| Containerfile | Писать отчёты в оговорённый путь внутри stage (напр. `/out/junit.xml`, coverage) и иметь export-stage **или** поддерживать export через build `--output` |
| coin-executor | После test выгрузить отчёты в workspace **`.coin/test-results/`** (BuildKit/podman `--output type=local,dest=.coin/test-results` с target вроде `test-reports`, либо эквивалент) |
| coin-lib | `archiveArtifacts artifacts: '.coin/test-results/**', allowEmptyArchive: true`; опционально `junit` pattern для XML |

Пример Containerfile (ориентир для seed/fixture):

```dockerfile
FROM golang:1.22 AS test
WORKDIR /src
COPY . .
RUN mkdir -p /out && go test ./... -coverprofile=/out/coverage.out

FROM scratch AS test-reports
COPY --from=test /out/ /
```

Падение тестов: executor MUST по возможности всё равно выгрузить уже записанные отчёты (best-effort), затем fail task. Точная механика fail+export — реализация; контракт директории `.coin/test-results/` стабилен.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| `resolve: file` утечёт в product main | Soft warn; docs «platform/samples»; hard guard — отдельный follow-up |
| Fixture drift от будущего API materialize | Один JSON Schema v4; control-plane change MUST match fixture shape |
| Breaking v3 remote path | Этот change не требует v4 remote; remote — `pipeline-v4-control-plane` |
| UI/API scope creep | Вынесено в `pipeline-v4-control-plane` |
| `sh` security | Allowlist в executor |
| Test reports lost on failed `RUN` | Export stage / best-effort extract; document Containerfile pattern |

## Migration Plan

1. ADR `docs/adr/pipeline-tekton-mapping.md`.
2. JSON Schema `pipeline-inline.v4.schema.json` (+ обновление `manifest.schema.json` / config.v2).
3. config v2: `resolve`, `manifestFile`; coin-lib file resolve + soft warn.
4. coin-executor: v4 parse, step dispatch, test-via-target + export `.coin/test-results/`, `--task`.
5. coin-lib: `pipeline.tasks` → stages, `--task`, archiveArtifacts test-results.
6. Sample/demo: `.coin/manifest.local.yaml` + `resolve: file`; Containerfile с `test` / `test-reports`; docs.
7. Acceptance: validate → test → build (publish по branching) через **file resolve**, без API/UI.

API/UI/remote E2E — [`pipeline-v4-control-plane`](../pipeline-v4-control-plane/).

Rollback: убрать `resolve: file` из config; remote path прежний.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | `sh` step / cmd на агенте для test? | ✅ | A: нет, только containerfile target / B: allowlisted sh | **A** (D12); `sh` не для default test |
| Q10 | Test reports path | ✅ | `.coin/test-results/` + archiveArtifacts | **принято** (D12) |
| Q11 | Testcontainers / run+socket | ⏸ later | вне Phase A | spike отдельным follow-up |
| Q2 | `when` на task vs branching-only | ✅ | A: task-level `when` / B: только branching | **A** (сохранить v3 semantics) |
| Q3 | UI graph library | → control-plane | см. `pipeline-v4-control-plane` | — |
| Q4 | File resolve guardrail | ✅ | Soft / Hard | **Soft** warn |
| Q5 | Default `manifestFile` | ✅ | `.coin/manifest.local.yaml` | **принято** |
| Q6 | YAML vs JSON fixture | ✅ | A: YAML (human) / B: JSON only | **A** default; lib MAY also accept JSON |
| Q12 | `kind: sh` в Phase A runtime? | ✅ | A: fail-closed / B: allowlist / C: omit | **A** — execute reject (не default test) |
| Q13 | Agent в fixture | ✅ | dual podContainers (драфт) / single `runtime.image` | **single** per coin-ci-runtime; dual — вне scope |
| Q14 | Кто валидирует v4 в Phase A? | ✅ | A: validate на load в executor / B: только тесты | **A** минимальный validate на load |
| Q7 | Имена kind containerfile | ✅ | A: project+coin / B: project+managed | **B** (`managed`\|`project`) |
| Q8 | imageRef в build/publish step | ✅ | A: литерал в YAML / B: computed + outputs | **B** (D11) |
| Q9 | N build ↔ N publish | ✅ | один outputs.json merge | **принято** |
| Q15 | Destinations: имя секции / multi-push / cache | ✅ | catalog `destinations[]`; publish `destinationRefs[]`; build `cache` override | **принято** (D13); demo — один `nexus-docker` |
