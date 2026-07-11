## Why

После IA-рефакторинга (coin-ui-enabling-ia) платформенные компоненты разнесены по трём каталогам: Runtime, Build stacks, Branching models. Страница `/platform/components` осталась как deprecated aggregate — дублирует три каталога, путает enabling team и сохраняет generic create flow вне family hubs.

Сейчас — момент убрать: family-каталоги покрывают все сценарии, в sidebar aggregate не отображается.

## What Changes

- **BREAKING:** Удалить маршрут `/platform/components` и страницу `PlatformComponentsPage`.
- Перенаправить legacy bookmarks `/components` и `/platform/components` на `/platform/runtime` (дефолтный Platform-каталог).
- Убрать footer-ссылку «Legacy: all components» с family catalog pages.
- Расширить `LegacyComponentDetailRedirect`: `/components/:type/:name` → family hub для известных типов (`agent`, `gp-content`, `branching-model`).
- Обновить docs (`coin-ui-user-guide.md`, `coin-ui/README.md`).

## Non-goals

- Удаление `/components/:type/:name/:version` (`ComponentDetail`) — отдельный legacy path, вне scope.
- Изменения coin-api или registry API.
- Редирект `executor` компонентов (нет family hub в UI).

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `ui-enabling-shell`: обновить legacy redirect для `/components`; убрать требование aggregate view на `/platform/components`.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-ui** | `App.tsx`, удаление `PlatformComponentsPage.tsx`, `PlatformProfileCatalogPage.tsx` |
| **docs** | user guide, coin-ui README |
| **openspec** | delta `ui-enabling-shell` |
