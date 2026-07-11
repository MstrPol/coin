## Why

Build stack (`gp-content`) редактируется частичной формой и line-based YAML parser, тогда как branching model уже прошёл путь **schema v2 → validate → preview API → bijective visual editor**. Эталонные `content.yaml` содержат legacy (`controls`, `when: tag`), три build engine (включая redundant `dockerfile` managed path и buildpack), а продуктовый контракт config v2 не допускает project-level build overrides.

После синхронизации ADR ([sync-runtime-docs-adr](../archive/2026-06-29-sync-runtime-docs-adr/)) нужен согласованный **gp-content schema v2**, двухengine runtime и конструктор build stack по образцу branching.

## What Changes

- **BREAKING:** `content.yaml` только `schemaVersion: 2`; v1 без `schemaVersion` → reject в validate-package.
- **BREAKING:** Два build engine: `buildkit` (platform managed Containerfile) и `dockerfile` (BYO Dockerfile в репо продукта; политика пути в GP `content.yaml`, **не** в `config.yaml`).
- **BREAKING:** Hard cut **buildpack** — удалить engine, `go-app-bp`, `pack`, `paketo-builder.tar`, E2E buildpack job.
- **BREAKING:** Удалить `go-app-df` и старую семантику `dockerfile` engine (managed Containerfile duplicate buildkit-lite).
- Убрать из schema: `controls`, `pipeline.stages[].when` (publish gate — branching + Jenkins `params.publish`).
- Добавить `capabilities.deliverables`: `artifact` только для `buildkit`; BYO `dockerfile` GP — `image` only.
- `gp-content.schema.json` v2; эталон `go-app` + новый BYO GP (напр. `go-app-docker`).
- **coin-api:** строгий validate-package; `POST /v1/admin/gp-content/preview` (manifest subset + warnings).
- **coin-ui:** `GpContentEditor` cards (engine, targets, pipeline, artifacts, capabilities); bijective `gpContentYaml.ts`; preview panel.
- **coin-executor:** BYO dockerfile из workspace; убрать buildpack dispatch; убрать materialize Containerfile для BYO engine.
- **coin-agent:** убрать `pack` и buildpack bootstrap.
- Amend ADR [build-engine-contract](../../docs/adr/build-engine-contract.md), [coin-ci-runtime](../../docs/adr/coin-ci-runtime.md).

## Capabilities

### New Capabilities

- `gp-content-preview`: admin preview API для draft gp-content (resolved `build` + `pipeline` fragment, engine-specific warnings).

### Modified Capabilities

- `platform-build-stacks`: schema v2 editor (2 engines), capabilities, artifacts tabs, preview panel.
- `build-engine`: два engine; REMOVED buildpack; BYO dockerfile semantics; E2E 2 engines; artifact только buildkit.
- `runtime-documentation`: doc matrix 2 engines + BYO GP model.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-gp-content** | schema v2, go-app migrate, go-app-bp/df delete, go-app-docker add |
| **coin-executor** | buildpack remove, BYO dockerfile path |
| **coin-api** | validate v2, preview endpoint, manifest builder |
| **coin-ui** | GpContentEditor rewrite, api client |
| **coin-lib** | bootstrap без buildpack load |
| **Dockerfile.agent** | без pack/paketo tar |
| **docker/scripts** | seed, e2e 2 engines |
| **docs** | agent-build-model, how-to build stacks |

## Non-goals

- `project.dockerfile` или иные build поля в `config.yaml`.
- Buildpack engine (hard cut, не feature flag).
- Три engine в UI/schema.
- Managed Containerfile для `dockerfile` engine (старый go-app-df).
- Fleet corp rollout / wave migration.
