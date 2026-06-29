# trunk-based (schema v2)

Сервисные Golden Path: `go-app`, `go-app-bp`, `go-app-df`.

## Правила

| # | name | branch example | template | publish |
|---|------|----------------|----------|---------|
| 1 | main | `main` | `v{base}-main-snapshot-{n}` | false |
| 2 | feature | `feature/PROJ-101` | `v{base}-{jira}-snapshot-{n}` | false |
| 3 | bugfix | `bugfix/PROJ-101` | `v{base}-{jira}-snapshot-{n}` | false |
| 4 | release | `release/PROJ-404` | `v{base}-{jira}-rc-{n}` | true |

Первое совпавшее правило выигрывает. Regex — Go RE2, named groups `(?P<jira>...)`.

Publish: Jenkins `params.publish=true` → `COIN_PUBLISH_REQUEST=true`; ветка с `publish: false` → **FAIL** stage publish.

См. [docs/how-to/branching-models.md](../../docs/how-to/branching-models.md).
