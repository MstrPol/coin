## 1. coin-ui — routes и redirects

- [x] 1.1 Удалить route `platform/components` и import `PlatformComponentsPage` из `App.tsx`
- [x] 1.2 Redirect `/components` → `/platform/runtime` (вместо `/platform/components`)
- [x] 1.3 Redirect `/platform/components` → `/platform/runtime`
- [x] 1.4 Расширить `LegacyComponentDetailRedirect`: `platformHubPath` для agent / gp-content / branching-model
- [x] 1.5 Удалить файл `PlatformComponentsPage.tsx`

## 2. coin-ui — cleanup UI

- [x] 2.1 Убрать footer «Legacy: all components» из `PlatformProfileCatalogPage.tsx`

## 3. Docs

- [x] 3.1 Обновить `docs/coin-ui-user-guide.md` — убрать `/platform/components` из redirects/legacy
- [x] 3.2 Обновить `coin-ui/README.md` — убрать строку `/platform/components`

## 4. Validation

- [x] 4.1 `npm run build` в coin-ui
- [x] 4.2 `openspec validate remove-platform-components-legacy --strict`
