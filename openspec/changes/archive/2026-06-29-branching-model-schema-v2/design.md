## Context

**Текущее состояние (schema v1):**

```yaml
schemaVersion: 1
trunk: { branch: main }
branchTypes: [feature, bugfix, release]
versioning: { tagPrefix, qualifiers: { snapshot, rc } }
publish: { when: tag | branch | ... }
```

- Executor: фиксированный regex `<type>/<JIRA-ID>`, `semver-tag` по `name == "semver-tag"`.
- Publish: auto policy по тегу (`ShouldPublish`).
- UI: плоская форма `BranchingModelEditor`.
- Docs: monolith `docs/branching.md`.

**Согласованная модель (explore, 2026-06):**

- Hard cut v2, без потребителей v1.
- Branch-centric rules, first match wins.
- Versioning: mini-DSL `template` (не prefix/suffix).
- Publish: двухуровневый — Jenkins param + `branches[].publish` eligibility.
- `publish=true` + недопустимая ветка → **fail**.

## Goals / Non-Goals

**Goals:**

- Единый согласованный `model.yaml` v2 как SoT для конструктора и executor.
- Preview API = executor SoT (не дублировать в TypeScript).
- Operator docs: how-to редактора + per-model cheat sheets.
- E2E green с v2 `trunk-based` на demo-go-app.

**Non-goals:**

- schema v1 adapter.
- Свободный template DSL beyond documented placeholders.
- Jenkins multibranch include/exclude rules в YAML.

## Decisions

### D1. schemaVersion 2 hard cut

**Решение:** только v2 в validate, UI, seed. Удалить v1 schema, structs, client types.

**Альтернатива:** adapter v1→v2 — отклонено (нет потребителей).

### D2. YAML shape

```yaml
schemaVersion: 2
name: trunk-based          # == component profile name

branches:                  # ordered; first match wins
  - name: main             # rule id (UI label)
    pattern: ^main$|^master$
    versioning:
      template: "v{base}-main-snapshot-{n}"
    publish: false         # eligibility when publish requested

  - name: release
    pattern: ^release/(?<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$
    versioning:
      template: "v{base}-{jira}-rc-{n}"
    publish: true
```

- Нет отдельного `trunk:` — `main`/`master` как первая карточка.
- `pattern` — RE2; named captures → template placeholders.
- Уникальные `branches[].name` в пределах модели.

### D3. Template mini-DSL

**Placeholders (v2 pilot):**

| Token | Источник |
|-------|----------|
| `{base}` | semver major.minor.patch из tag history |
| `{jira}` | named capture `jira` |
| `{n}` | increment в серии тегов с тем же prefix |
| `{branch}` | rule `name` |
| литералы | `v`, `-rc-`, `-snapshot-` в строке |

`COIN_VERSION` = tag body без ведущего литерала `v` если template начинается с `v`.

Validate: `{jira}` в template требует `(?<jira>...)` в pattern; `{n}` требует серийного шаблона.

### D4. Publish — param + eligibility

```
Layer 1 (coin-lib):  params.publish == false → skip publish stage
Layer 2 (executor):  COIN_PUBLISH_REQUEST == true (set by coin-lib when param true)
                     + matched rule.publish == true → run publish
                     + COIN_PUBLISH_REQUEST == true + rule.publish == false → FAIL
```

**Решение:** executor не auto-publishes по тегу. Тег влияет на `ResolveVersion`, не на trigger publish.

**Альтернатива:** skip с exit 0 — отклонено пользователем → **fail**.

Env: `COIN_PUBLISH_REQUEST=true|false` (coin-lib sets from `params.publish`).

### D5. Branch match — first wins

Порядок `branches[]` в YAML = приоритет. UI: reorder cards (up/down).

No match → `ValidateBranch` error.

### D6. Pull Request builds

PR не отдельный pattern. coin-lib нормализует ветку для Coin:

- Prefer `CHANGE_BRANCH` when set (PR multibranch).
- Иначе `GIT_BRANCH` / `BRANCH_NAME`.

PR на `feature/PROJ-101` → feature rule → `publish: false`.

### D7. Preview API (отдельный endpoint)

```
POST /v1/admin/branching-models/preview
```

Body: inline `model` (v2) + `scenarios[]` with `branch`, `tagName?`, `tags?`, `requestPublish?`.

Response: per scenario `matchedRule`, `branchValid`, `coinVersion`, `publishAllowed`, `publishOutcome` (allowed | denied | not-requested), errors.

Не расширять `validate-package`.

### D8. manifest.branching

Resolve копирует `branches[]` as-is + pin `name` / `version`. Без отдельного materializer shape.

### D9. Docs

- Delete `docs/branching.md` (no stub).
- Add `docs/how-to/branching-models.md`.
- Compress `coin-branching-models/models/*/README.md` (cheat sheet + scenarios table for UI presets).

### D10. UI constructor

Bijective editor for v2 YAML:

- Ordered branch cards: name, pattern, template, publish toggle.
- Test branch name + patternHint from preview API.
- Right panel: preset + custom scenarios with `requestPublish` toggle.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Invalid RE2 in pattern | compile at validate; show error in UI |
| coin-api imports coin-executor | thin `branching/preview` package |
| Breaking local seed data | re-run seed-jenkins-lib |
| Partial draft changes `branching-docs-catalog` / `branching-editor-preview` | superseded by this change; remove orphan dirs |

## Migration Plan

1. schema v2 + executor rewrite + tests.
2. coin-api validate + preview + OpenAPI.
3. coin-lib `COIN_PUBLISH_REQUEST`.
4. coin-ui constructor + preview.
5. Migrate model.yaml catalog + seed scripts.
6. Docs hard cut.
7. `openspec validate --strict`; E2E demo-go-app.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| — | — | — | Нет blocking вопросов |
