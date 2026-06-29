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

## Два уровня publish

1. Jenkins `params.publish` → coin-lib выставляет `COIN_PUBLISH_REQUEST=true`.
2. `branches[].publish` — eligibility по совпавшему правилу.

`requestPublish=true` + `publish: false` на ветке → **FAIL** stage publish (не skip).

## Каталог

Эталоны: [`coin-branching-models/models/`](../coin-branching-models/models/).

GP pin: branching-model slot в composition → `manifest.branching` при resolve.

## UI и preview

Component Studio: rule builder + `POST /v1/admin/branching-models/preview` (executor SoT).

## См. также

- [ADR: GP branching model](../adr/gp-branching-model.md)
- [coin-branching-models README](../coin-branching-models/README.md)
