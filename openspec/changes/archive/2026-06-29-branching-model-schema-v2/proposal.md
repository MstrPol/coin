## Why

Текущий `model.yaml` schema v1 (глобальный `trunk`, `branchTypes`, `publish.when`) не отражает согласованную модель: **список branch rules** с regex, template DSL и per-branch publish eligibility. Executor использует захардкоженные паттерны и выбор стратегии по имени модели. Enabling team не может честно собрать кастомные модели в UI, а `docs/branching.md` описывает одну глобальную политику.

Потребителей schema v1 в prod нет — делаем **hard cut** на schema v2, визуальный конструктор, preview API и обновление docs в одном change.

См. ADR [gp-branching-model.md](../../docs/adr/gp-branching-model.md).

## What Changes

- **BREAKING:** `model.yaml` только `schemaVersion: 2` — формат `branches[]` с `pattern`, `versioning.template`, `publish` (eligibility).
- **BREAKING:** Удалить schema v1, adapter и v1 UI fields (`trunk`, `branchTypes`, `qualifiers`, `publish.when`).
- **coin-executor:** новый branching engine — first-match rules, template DSL, `AllowsPublish(branch)`.
- **Publish semantics:** intent = Jenkins `params.publish`; модель = допуск по ветке; `publish=true` + `publish: false` на rule → **fail** (не skip).
- **coin-api:** `POST /v1/admin/branching-models/preview` (отдельный от validate-package).
- **coin-ui:** rule builder (карточки веток) + scenario preview + test branch name.
- **coin-lib:** передать `COIN_PUBLISH_REQUEST=true` при `params.publish=true` для executor guard.
- **Docs:** удалить `docs/branching.md`; `docs/how-to/branching-models.md` + сжатые model READMEs.
- **Seed/E2E:** эталонные модели `trunk-based`, `semver-tag` на v2 YAML.

## Capabilities

### New Capabilities

- `branching-model-preview`: admin preview API с executor-backed scenario evaluation.

### Modified Capabilities

- `branching-model`: schema v2 only; model schema requirement.
- `executor-branching`: branch rules match, template versioning, publish eligibility + fail on denied request.
- `branching-models-catalog`: rule builder UI, preview panel, docs links.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-branching-models** | schema v2, model.yaml, READMEs |
| **coin-executor** | rewrite `internal/branching` |
| **coin-api** | validate v2, preview endpoint, OpenAPI |
| **coin-ui** | constructor, api client |
| **coin-lib** | `COIN_PUBLISH_REQUEST` env |
| **docs** | delete branching.md, how-to, link fixes |
| **docker/scripts** | seed, e2e-branching-policy |

## Non-goals

- Обратная совместимость schema v1 (нет adapter).
- Настраиваемый regex DSL вне RE2 `pattern` на карточке.
- Per-repo branching overrides.
- Отдельный branch type для Pull Request (`PR-N`); PR = source branch pattern.
- Fleet migration wave / corp rollout.
- `coin version bump` CLI GA (best-effort в preview optional).
