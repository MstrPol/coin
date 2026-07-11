## Context

`rework-build-stack-contract` (архив `2026-07-03-rework-build-stack-contract`) реализовал Build Stack vNext с каталогами targets, deliverables, containerfiles и pipeline по id. Platform lead выбрал **вариант C** (inline steps) и уточнил: **Containerfile не выносить в отдельную секцию** — на pilot контент сборки живёт **вместе с шагом stage**, и resolved manifest отражает это же.

Ограничения без изменений: product config v2, build logic в executor, coin-lib glue only, parameters non-secret, local pilot hard cut.

## Goals / Non-Goals

**Goals:**

- Author-facing модель: **Parameters → Pipeline stages** (две секции).
- Каждый `run` / `build` step с `engine: buildkit` содержит inline `containerfile.body` (или materialized `contentRef` в manifest на том же step).
- Каждый step с `engine: dockerfile` содержит inline `dockerfile.path` (BYO из workspace продукта).
- Resolved manifest самодостаточен для Nexus fallback: executor читает step + containerfile с одного объекта.
- coin-ui: pipeline-first, без catalog cards и без отдельной секции Containerfiles.

**Non-goals:**

- Отдельный каталог `artifacts.containerfiles` в author model и top-level manifest (pilot).
- DRY-оптимизация shared containerfile между steps (допустимо дублирование body на pilot; dedupe только внутри Nexus package storage).
- Corp fleet rollout, YAML export/debug.

## Decisions

### Решение 1: `schemaVersion: 3`, только parameters + pipeline + validateSchema

```yaml
schemaVersion: 3
name: go-app
kind: gp-content

parameters:
  - name: GO_VERSION
    type: string
    default: "1.22"
    required: true

validateSchema: schemas/config.v2.schema.json

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
```

**Отдельной секции `artifacts.containerfiles` нет.** Единственный вне-pipeline артефакт — `validateSchema` (schema product config).

### Решение 2: Typed step actions (без изменений)

| action | Inline block |
|--------|----------------|
| `run` | `run: { engine, output?, containerfile?, dockerfile? }` — стадия Containerfile выводится из `output` |
| `build` | `build: { id, type, engine, containerfile?, dockerfile?, image?, artifact? }` — стадия из `type` |
| `publish` | `publish: { buildStepId }` |

Явные поля `target` / `dockerfile.target` в author model **не используются** (pilot): executor выводит multi-stage имя из `run.output` или `build.type`.

`build.id` уникален в stack.

### Решение 2a: Short hash ids (`^[a-z0-9]{5,6}$`)

Machine ids для `pipeline.stages[].id` и `build.id` — **короткий hash 5–6 символов** (lowercase alphanumeric). UI генерирует id при создании stage/build step; отображаемое имя — в `stage.name`. coin-api validate-package отклоняет id вне формата.

### Решение 3: Containerfile inline в buildkit step (pilot)

Для `engine: buildkit` step **обязан** содержать:

```yaml
containerfile:
  body: "<полный текст Containerfile>"
```

- Author model и gp-content package хранят `body` на step.
- Manifest materializer добавляет на **тот же step** resolved поля:

```yaml
containerfile:
  contentRef: "nexus://..."
  digest: "sha256:..."
```

Top-level `artifacts.containerfiles[]` в manifest **не создаётся** на pilot.

**Пилот go-app:** допустимо повторять один и тот же `body` в нескольких steps (validate, test, build). Materializer MAY dedupe blobs в package storage, но manifest preview показывает containerfile co-located с каждым step.

**Альтернатива (отклонена):** named catalog `artifacts.containerfiles` + `containerfile: app` ref — воспроизводит отделение, которое отверг platform lead.

### Решение 4: BYO dockerfile inline в step

Для `engine: dockerfile`:

```yaml
dockerfile:
  path: Dockerfile   # путь в workspace продукта
  target: runtime    # optional
```

Без managed containerfile body.

### Решение 5: Resolved manifest = pipeline-inline + per-step containerfile

Manifest содержит:
- `parameters`
- `validateSchema` (ref)
- `pipeline.stages[]` с inline steps; у buildkit steps — `containerfile` с `contentRef`/`digest` на том же объекте

Нет: `build.targets`, `deliverables`, `artifacts.containerfiles`.

Executor materializes Containerfile из step.containerfile перед run/build dispatch.

### Решение 6: UI — только Parameters и Pipeline

1. Parameters
2. Pipeline stages (primary) — в buildkit step: textarea Containerfile body inline в карточке step
3. Resolved manifest preview (справа)

### Решение 7: Migration go-app v2 → v3

| v2 | v3 |
|----|-----|
| `artifacts.containerfiles[0].body` + targets | `containerfile.body` скопирован в каждый соответствующий run/build step |
| `run-target` | `action: run` + inline `run` |
| `build-deliverable` | `action: build` + inline `build` |
| `publish-deliverable` | `action: publish` + `buildStepId` |

## Risks / Trade-offs

- **Дублирование containerfile body между steps** → принято для pilot; позже можно ADD optional `inheritFromStep` без отдельного каталога.
- **Большой manifest** → contentRef + digest на step, не полный body в resolved manifest (только ref).
- **Breaking v2/vNext drafts** → hard cut + reseed.

## Migration Plan

Без изменений по фазам; убрать задачи на отдельную Containerfiles UI секцию.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | `schemaVersion: 3` hard cut | ✅ decided | A: v3; B: flag | A |
| Q2 | Publish ref | ✅ decided | `buildStepId` | A |
| Q3 | Containerfile placement | ✅ decided | A: отдельный catalog; B: inline в step | B: inline в step + manifest co-located |
| Q4 | Resolved manifest: body vs ref | ✅ decided | A: full body; B: contentRef на step | B: contentRef+digest на step (self-sufficient resolve) |
| Q5 | Dedupe одинаковых body | ✅ decided | A: обязателен; B: optional internal | B: optional dedupe в package storage only |
