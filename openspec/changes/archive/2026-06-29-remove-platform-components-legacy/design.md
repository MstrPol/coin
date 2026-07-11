## Context

coin-ui Platform IA (ADR: enabling shell) ввела три family-каталога:

| Family | Route | compType |
|--------|-------|----------|
| Runtime | `/platform/runtime` | `agent` |
| Build stacks | `/platform/build-stacks` | `gp-content` |
| Branching models | `/platform/branching-models` | `branching-model` |

`/platform/components` — aggregate всех типов с generic create form. В nav скрыт с 2026-06, но route и redirect `/components` → `/platform/components` остались. Footer family-каталогов ссылается на aggregate.

См. archived change `coin-ui-enabling-ia`, spec `ui-enabling-shell`.

## Goals / Non-Goals

**Goals:**

- Hard cut: нет страницы aggregate, нет ссылок на неё.
- Bookmarks `/components` и `/platform/components` ведут на `/platform/runtime`.
- `/components/agent/:name` и аналоги для gp-content / branching-model → family hub.
- Spec и docs синхронизированы.

**Non-goals:**

- Удаление `ComponentDetail` (`/components/:type/:name/:version`).
- Редирект `executor` / `lib` типов.
- E2E browser tests (достаточно `npm run build`).

## Decisions

### D1. Redirect target для `/components` и `/platform/components`

**Решение:** `/platform/runtime` — первый пункт Platform nav, agent stacks — наиболее частый entry.

**Альтернатива:** dashboard `/` — хуже, теряется контекст Platform.

### D2. Удалить файл, не deprecate label

**Решение:** удалить `PlatformComponentsPage.tsx` целиком. `ComponentCatalogTable` остаётся — используется family-каталогами через `PlatformCatalogPage` / profile pages.

### D3. Legacy component detail redirect

**Решение:** в `LegacyComponentDetailRedirect` использовать `platformHubPath(type, name)` из `platformFamilyConfig` для всех типов с family mapping; fallback — существующий `ComponentDetail`.

```tsx
const hub = platformHubPath(type, name);
if (hub) return <Navigate to={hub} replace />;
return <ComponentDetail />;
```

### D4. Footer link removal

Убрать блок «Legacy: all components» из `PlatformProfileCatalogPage` — без замены.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Bookmarks на `/platform/components` | Redirect на `/platform/runtime` |
| Generic create flow потерян | Create доступен в каждом family catalog (`New profile`) |
| `/components/executor/...` без hub | Остаётся `ComponentDetail` (non-goal) |

## Migration Plan

1. Deploy coin-ui с redirects до удаления страницы (один PR).
2. Обновить docs.
3. `openspec validate --strict`.

Rollback: revert PR (route + page восстанавливаются).

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| — | — | — | Нет blocking вопросов |
