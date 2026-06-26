## Why

После `platform-native-lifecycle` Platform-каталоги (`/platform/runtime`, `/platform/build-stacks`, `/platform/branching-models`) остались **плоскими списками версий** без entity hub, как у GP (`/gp/{name}`). Operator теряет контекст профиля компонента, кнопки create непоследовательны («Create draft» vs отсутствие create), а runtime намеренно исключён из draft lifecycle — хотя Jenkins-пайплайн уже публикует образ до регистрации в API и нужен полный цикл draft → published с ручным catch-up в UI.

Сейчас: `publish-agent.sh` регистрирует `agent/coin-agent@version` **сразу как published**; branching-models catalog без create; build-stacks и GP hub используют разную терминологию и навигацию. Модель «профиль компонента + релизы (версии)» не доведена до Platform.

## What Changes

- **Platform Component Hub** — GP-like shell для трёх семейств: agent stacks (runtime), build stacks (`gp-content`), branching models.
- Каталоги (`/platform/{family}`) показывают **профили** (имена компонентов), не плоский список всех версий.
- Hub URL: `/platform/runtime/{name}`, `/platform/build-stacks/{name}`, `/platform/branching-models/{name}` с вкладками Overview + Releases (и editor routes для draft).
- Унификация primary action: **«New draft»** на hub; **«New profile»** на каталоге семейства (аналог `/gp/new`).
- **BREAKING (spec)**: Runtime (`agent`) получает полный lifecycle **draft → published**; coin-ui предоставляет draft create/promote/delete и release detail (metadata image/digest).
- API: register agent version как `draft` (Jenkins после push image) + promote в `published`; идемпотентность для CI; миграция `publish-agent.sh`.
- Release detail для agent: read-only derived pin `executor/coin-executor@{version}` (не отдельный релиз в hub).
- Задел на несколько agent profiles: `coin-agent`, `coin-agent-arm`, `coin-agent-minimal`.
- Redirects: flat version URLs и старые catalog deep links → hub hierarchy.
- **Supersedes** часть `platform-component-lifecycle`: requirement «Runtime components published only» и «No draft runtime versions in UI».

### Non-goals

- UI authoring Docker images / Dockerfile для agent (build остаётся script/CI-first).
- Отдельный hub tab или lifecycle для `executor` (производный pin от agent stack).
- Auto-lock draft agent pins при назначении на GP canary line (открытый вопрос — см. design).
- Wave rollout / corp fleet migration.
- Component-level canary (остаётся draft/published only).

## Capabilities

### New Capabilities

- `platform-component-hub`: единый hub pattern (layout, tabs, routes, create actions) для agent, gp-content, branching-model; shared terminology и redirects.

### Modified Capabilities

- `platform-component-lifecycle`: agent включается в two-state lifecycle; убрать runtime exception (published only); уточнить Jenkins register → draft → promote path.
- `platform-runtime-catalog`: каталог профилей agent stack; hub navigation; draft/publish UI вместо script-only guidance.
- `platform-build-stacks`: каталог профилей + hub; унификация «New draft» / «New profile»; redirects с flat URLs.
- `branching-models-catalog`: create на каталоге; hub navigation; унификация actions.
- `gp-entity-hub`: ссылки из GP release composition на platform hub URLs (не flat catalog).

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-ui** | `PlatformComponentHubLayout`, hub pages для 3 families; refactor `PlatformCatalogPage`; routes в `App.tsx`; унификация labels |
| **coin-api** | Admin endpoints: agent draft register, promote; optional list-by-profile; идемпотентность POST |
| **coin-executor** | `publish-agent.sh`: register draft + promote (или split steps) |
| **docs** | runbook agent publish, coin-ui-user-guide |
| **E2E** | hub navigation, agent draft lifecycle, create branching model |
| **OpenSpec** | новый spec + 5 delta specs; amend ADR при необходимости |
