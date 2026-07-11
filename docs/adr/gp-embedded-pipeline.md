# ADR: GP embedded pipeline

**Статус:** accepted  
**Дата:** 2026-07-03  
**Связанный change:** `gp-release-embedded-pipeline`  
**Supersedes:** gp-content as platform component ([gp-component-package-model.md](./gp-component-package-model.md) — секция gp-content)

## Контекст

После `pipeline-inline-steps` build contract — parameters + pipeline stages с inline containerfile. Отдельный platform component `gp-content` создаёт двойной promote, разрыв UX и legacy reuse (`xxx → go-app`). Продукт pin'ит только GP release (`coin.goldenPath` + `coin.version`).

## Решение

### GP release = primary entity

Pipeline-inline `schemaVersion: 3` хранится как **embedded body** GP release draft (`gp_release_pipeline_bodies`). Published SoT — materialized внутри Nexus manifest blob при GP promote.

### Composition — 2 pin

| Slot | Type |
|------|------|
| `agent` | `agent` |
| `branching-model` | `branching-model` |

`gp-content` slot и component type **удалены**.

### Identity

`gp_profiles.name` = pipeline family = `coin.goldenPath`. Reuse pipeline между профилями **не поддерживается**.

### Authoring

Enabling team редактирует pipeline на GP release detail (Pipeline section). Platform → Build stacks **удалён**.

### Preview API

`POST /v1/admin/golden-paths/{name}/versions/{version}/pipeline/preview`

### coin-gp-content

Git seed source для bootstrap (`stacks/*`); `publish-content.sh` — не primary path.

## Последствия

- coin-api: `gp_release_pipeline_bodies`, resolve без gp-content package
- coin-ui: pipeline editor на GP release detail
- coin-executor: без изменений manifest shape
- Hard cut local pilot: wipe/reseed допустим

## Отклонённые альтернативы

| Альтернатива | Почему нет |
|--------------|------------|
| Скрыть Build stacks в UI, оставить registry | Двойная модель остаётся |
| gp-content reuse между GP profiles | Platform lead: не нужен |
| Отдельный semver pipeline | Один cadence = GP release version |
