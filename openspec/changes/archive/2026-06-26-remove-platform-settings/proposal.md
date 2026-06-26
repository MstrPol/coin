## Why

После рефакторинга Platform (runtime / build-stacks / branching-models с hub и draft→publish) страница `/platform-settings` вводит в заблуждение: название ассоциируется с настройками platform-компонентов, хотя там только singleton Nexus (`nexusMavenBase`, `nexusCredentialsId`). Эти значения **не читаются** runtime coin-api — реальный SoT для Nexus уже **env** (`NEXUS_URL`, `NEXUS_MAVEN_RELEASES`, …). UI и API — мёртвый слой с audit noise.

## What Changes

- **BREAKING**: Удалить `GET/PUT /v1/admin/platform/settings` из coin-api.
- **BREAKING**: Удалить `/platform-settings` route и nav entry из coin-ui.
- Migration: drop таблица `platform_settings`.
- Удалить `PlatformSettings.tsx`, типы и client methods.
- OpenAPI: убрать `PlatformSettings` schemas и paths.
- Docker seed/E2E: убрать `PUT /platform/settings`; Nexus — только через compose `.env`.
- Redirect `/platform-settings` → `/audit` (bookmarks).
- Документация: SoT Nexus = env/runbook, не operator UI.

### Non-goals

- Corp Helm chart для Nexus (отложено до corp gate).
- Новая страница «Integrations» / замена UI (вариант B/C отклонён).
- Изменение env-модели coin-api или Jenkins credential binding.
- Удаление audit action `update_platform_settings` из исторических записей.

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `ui-enabling-shell`: убрать requirement Platform settings; Admin nav без Platform settings; redirect legacy URL.
- `platform-runtime-catalog`: убрать упоминания Platform settings из anti-lib-pin requirement.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | Удалить handlers, store, admin service; migration drop table |
| **coin-ui** | Удалить page, nav, api methods, types; redirect |
| **docker** | `seed-jenkins-lib-stack.sh` без platform settings PUT |
| **docs** | coin-ui-user-guide, README, openapi.md, golden-paths (stale) |
| **OpenSpec** | 2 delta specs |
