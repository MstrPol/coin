## Why

Change `rework-build-stack-contract` ввёл каталоги targets, deliverables, containerfiles и pipeline по id — это не совпадает с ментальной моделью CI: после **Parameters** настраивается **pipeline**, и в каждом stage сразу видно что собрать/опубликовать и как, **включая Containerfile**.

## What Changes

- **BREAKING**: author-facing модель v3 = `parameters` + `pipeline` + `validateSchema` only; нет `build.targets`, `deliverables`, `artifacts.containerfiles`.
- **BREAKING**: buildkit steps несут inline `containerfile.body`; manifest materializer кладёт `contentRef`/`digest` **на тот же step**, не в top-level catalog.
- **BREAKING**: typed steps `run` / `build` / `publish` с inline config.
- **BREAKING**: coin-ui — **Parameters → Pipeline stages**; Containerfile textarea внутри buildkit step card.
- Обновить coin-api, executor, gp-content pilot stacks под v3.

## Capabilities

### New Capabilities

- `pipeline-inline-model`: schema v3 — parameters, pipeline stages, containerfile inline в steps.
- `manifest-pipeline-inline`: manifest без top-level targets/deliverables/containerfiles catalog; per-step containerfile refs.
- `build-stack-pipeline-ui`: pipeline-first editor, containerfile в step.

### Modified Capabilities

- `platform-build-stacks`: pipeline-first editor requirements.
- `gp-content-preview`: preview v3 inline model.
- `build-engine`: dispatch из inline steps + per-step containerfile.

## Impact

- `coin-ui`: убрать Targets, Deliverables, Containerfiles cards.
- `coin-api` / `coin-executor` / `coin-gp-content`: v3 contract.
- ADR `build-stack-vnext-contract` → superseded.

## Non-goals

- Отдельный catalog managed containerfiles (pilot).
- Fleet rollout, YAML debug editor, secrets в parameters.
- Обязательная дедупликация containerfile между steps (optional internal only).
