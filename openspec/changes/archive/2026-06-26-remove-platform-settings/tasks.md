## 1. coin-api — remove platform settings

- [x] 1.1 Migration: `DROP TABLE platform_settings`
- [x] 1.2 Удалить `store/platform_settings.go`, admin/service handlers, server routes
- [x] 1.3 OpenAPI: убрать `/v1/admin/platform/settings` и schemas
- [x] 1.4 Обновить `coin-api/README.md` — Nexus SoT = env only

## 2. coin-ui — remove page and nav

- [x] 2.1 Удалить `PlatformSettings.tsx`, route, import в `App.tsx`
- [x] 2.2 Убрать `platformSettings` / `updatePlatformSettings` из `api.ts` и types
- [x] 2.3 Убрать nav entry из `nav.ts`
- [x] 2.4 Redirect `/platform-settings` → `/audit`

## 3. Docker и scripts

- [x] 3.1 Убрать `PUT /platform/settings` из `seed-jenkins-lib-stack.sh`
- [x] 3.2 Проверить bootstrap/e2e scripts на ссылки на platform settings API

## 4. Docs

- [x] 4.1 Обновить `coin-ui/README.md`, `docs/coin-ui-user-guide.md`
- [x] 4.2 Обновить `docs/openapi.md`, `docs/golden-paths.md` (stale platform_settings)

## 5. Validation

- [x] 5.1 `go test ./...` в coin-api
- [x] 5.2 `npm run build` в coin-ui
- [x] 5.3 `openspec validate remove-platform-settings --strict`
