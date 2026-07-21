# Branching models (schema v2)

Модель ветвления — platform component `branching-model`. Правила задаются в `model.yaml` (`schemaVersion: 2`) и попадают в resolved manifest как `branching.branches[]`.

## Формат model.yaml

```yaml
schemaVersion: 2
name: trunk-based
branches:           # ordered — first match wins
  - name: main
    pattern: ^main$|^master$
    versioning:
      template: "v{base}-main-snapshot-{n}"
    publish: false
```

- **pattern** — Go RE2; named groups: `(?P<jira>...)`.
- **template** — mini-DSL: `{base}`, `{jira}`, `{n}`, `{branch}` + литералы.
- **publish** — допуск публикации по ветке (не auto-publish).

JSON Schema (документация контракта): [`docs/schemas/branching-model.schema.json`](../schemas/branching-model.schema.json). Runtime-валидация — в coin-api (Go) и coin-ui.

## Два уровня publish

1. Jenkins `params.publish` → coin-lib выставляет `COIN_PUBLISH_REQUEST=true`.
2. `branches[].publish` — eligibility по совпавшему правилу.

`requestPublish=true` + `publish: false` на ветке → **FAIL** stage publish (не skip).

## Source of truth

Authoring и lifecycle: Platform hub → coin-api (PG drafts) → Nexus (`published`).

Local pilot seed/E2E fixtures: [`docker/testdata/branching-models/`](../../docker/testdata/branching-models/) (`trunk-based`, `semver-tag`). Seed: `docker/scripts/seed-branching-model.sh`.

GP pin: branching-model slot в composition → `manifest.branching` при resolve.

## UI

Platform hub (`/platform/branching-models`): rule builder + YAML/card editing. Draft lifecycle: validate → register → publish.

Runtime behavior (versioning and publish eligibility) is evaluated by `coin-executor` from `manifest.branching` during CI; `coin-api` does not expose a branching scenario preview endpoint.

**Удаление draft:** Releases tab → **Delete** на строке draft, или в редакторе (`/platform/branching-models/{name}/{version}/edit`) → **Delete draft** в lifecycle panel (publisher+). Published версии удалить нельзя.

## См. также

- [ADR: GP branching model](../adr/gp-branching-model.md)
