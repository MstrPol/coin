# ADR: Pipeline-inline Build Stack

**Статус:** accepted  
**Дата:** 2026-07-03  
**Связанный change:** `pipeline-inline-steps`  
**Supersedes:** [build-stack-vnext-contract.md](./build-stack-vnext-contract.md)

## Контекст

Build Stack vNext ввёл каталоги `build.targets`, `deliverables` и `artifacts.containerfiles`, связанные с pipeline только по id. Это не совпадает с ментальной моделью CI: после **Parameters** настраивается **pipeline**, и в каждом stage сразу видно что собрать/опубликовать и как — **включая Containerfile**.

## Решение

### Canonical model (`schemaVersion: 3`)

Author-facing gp-content содержит только:

- `parameters`
- `validateSchema`
- `pipeline.stages[].steps[]` с inline typed config

Отдельных секций `build.targets`, `deliverables`, `artifacts.containerfiles` **нет**.

### Pipeline step actions

| action | Назначение |
|--------|------------|
| `run` | validate / test / промежуточный target |
| `build` | materialize output (`build.id`, `type`, engine config) |
| `publish` | publish по `publish.buildStepId` → `build.id` |

### Containerfile inline в step (pilot)

Buildkit `run` / `build` steps содержат:

```yaml
containerfile:
  body: |
    FROM golang:1.22-bookworm AS base
    ...
```

Resolved manifest materializer добавляет на **тот же step**:

```yaml
containerfile:
  contentRef: { url, sha256 }
```

Top-level `artifacts.containerfiles` в manifest **не создаётся** на pilot.

### Manifest

Resolved manifest содержит `parameters`, `validateSchema`, `pipeline.stages` с inline steps и per-step containerfile refs. Executor dispatch читает steps напрямую.

## Последствия

- `coin-api`: schema v3, validation, preview, manifest builder.
- `coin-ui`: Parameters → Pipeline stages; containerfile textarea в buildkit step card.
- `coin-executor`: `run` / `build` / `publish` inline dispatch; materialize containerfile from executing step.
- `coin-lib`: без изменений (glue only).
- Hard cut: v2 vNext catalog drafts отклоняются; pilot stacks reseed на v3.

## Отклонённые альтернативы

| Альтернатива | Почему отклонена |
|--------------|------------------|
| Отдельный catalog Containerfiles | Platform lead: «каша», не совпадает с pipeline-first UX |
| Hidden catalog + pipeline-first UI only | Два представления одной модели; лишняя сложность |
| v2 vNext catalog + dual path | Hard cut local pilot |
