## Why

После `pipeline-inline-steps` build stack — это parameters + pipeline stages с inline containerfile, то есть **суть Golden Path release**. Отдельный platform component `gp-content` создаёт лишний слой: двойной promote, разрыв UX (GP release detail → Platform → Build stacks), несвязанные semver и legacy-паттерн reuse (`xxx → go-app`). Продукт уже pin'ит только GP release; authoring должен совпадать с этой моделью.

## What Changes

- **BREAKING:** Удалить `component type: gp-content` из platform registry, API и UI.
- **BREAKING:** Pipeline-inline v3 (`schemaVersion: 3`) становится **primary payload GP release draft**, не отдельного component package.
- **BREAKING:** GP composition сокращается с 3 pin до 2: `agent` + `branching-model` (pipeline — intrinsic к release).
- **BREAKING:** Убрать Platform → Build stacks catalog, routes и component lifecycle для gp-content.
- GP profile `name` = `coin.goldenPath` = pipeline family identity; reuse pipeline между профилями **не поддерживается**.
- Enabling team редактирует pipeline на **GP release detail** (вкладка Pipeline); один promote gate на GP release.
- Preview/validate API переносится с `gp-content/preview` на GP release scope.
- Seed/bootstrap: pipeline из `coin-gp-content/stacks/` материализуется в GP releases, не в component registry.
- Supersede ADR `gp-component-package-model` (секция gp-content) и связанные playbook в `docs/golden-paths.md`.

## Capabilities

### New Capabilities

- `gp-embedded-pipeline`: хранение, validate, preview и materialize pipeline-inline v3 как тела GP release draft/published.
- `gp-release-two-pin`: composition GP release — ровно два внешних pin (`agent`, `branching-model`); pipeline не pin.

### Modified Capabilities

- `gp-composition-two-slot`: supersede трёхpin-моделью; gp-content slot удаляется.
- `platform-build-stacks`: capability удаляется целиком.
- `gp-content-preview`: supersede preview на GP release scope.
- `pipeline-inline-model`: owner модели — GP release, не gp-content package.
- `build-stack-pipeline-ui`: редактор pipeline на GP release detail, не Platform build stacks.
- `gp-entity-hub`: release detail — primary authoring surface для pipeline.
- `manifest-pipeline-inline`: materializer читает pipeline с GP release body.
- `gp-publish-flows`: единый promote GP release (без предварительного promote gp-content).
- `platform-component-lifecycle`: убрать lifecycle paths для type `gp-content`.
- `gp-profile-metadata`: profile name = pipeline family identity; запрет decoupled alias.

## Non-goals

- Изменение product config v2 (`coin.goldenPath` + `coin.version`).
- Коллапс `branching-model` или `agent` в GP release (остаются shared platform components).
- Corp fleet rollout и migration существующих prod GP (local pilot hard cut).
- Изменение Jenkins glue scope (`coin-lib` — resolve, pod, creds, executor stages only).
- Dual path / поддержка legacy gp-content registry entries.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | `gp_artifact_bodies` / release pipeline API; resolve без gp-content pin; удалить component APIs для gp-content |
| **coin-ui** | Pipeline tab на GP release detail; удалить `/platform/build-stacks/*`; draft wizard — 2 picker'а |
| **coin-executor** | Без изменений контракта manifest (pipeline sections те же) |
| **coin-gp-content** | Git templates/seed only; `publish-content.sh` deprecate |
| **docker/seed** | Seed GP releases с embedded pipeline; убрать `xxx`/`gp-01-07` alias → go-app |
| **OpenAPI** | Новые GP pipeline endpoints; удалить gp-content admin paths |
| **docs/adr** | Новый ADR; supersede gp-content sections в существующих ADR |
| **openspec/specs** | 2 new + 10 modified/removed capabilities |
