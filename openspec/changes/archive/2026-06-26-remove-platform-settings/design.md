## Context

**Текущее состояние:**

- `platform_settings` (PG singleton): `nexus_maven_base`, `nexus_credentials_id`, `updated_at`.
- API: `GET/PUT /v1/admin/platform/settings`; UI: `/platform-settings` в Admin nav.
- coin-api Nexus client использует **только env**: `NEXUS_URL`, `NEXUS_MAVEN_RELEASES`, `NEXUS_ADMIN_USER`, `NEXUS_ADMIN_PASSWORD`.
- После `jenkins-lib-outside-platform` из settings убран `runtime.lib`; остались только Nexus fields.
- Platform IA: versioned components (hub), не singleton infra config.

**Решение пользователя:** вариант **A** — убрать UI + API.

## Goals / Non-Goals

**Goals:**

- Убрать ложную «Platform settings» из operator console.
- Единый SoT для Nexus в local pilot: `docker/.env` + `coin-api` env.
- Упростить coin-api (меньше dead code и таблицы).

**Non-Goals:**

- Operator UI для Nexus в corp (Helm/Vault — отдельный gate).
- Миграция historical audit rows.

## Decisions

### D1: Drop table + API (hard cut)

**Решение:** migration `DROP TABLE platform_settings`; удалить endpoints без deprecation period.

**Альтернатива:** 410 Gone на API — отклонено; pilot fleet маленький.

### D2: Nexus SoT = environment

**Решение:** документировать в `coin-api/README.md` и `docker/README.md` — единственный путь настройки Nexus для control plane.

| Concern | SoT |
|---------|-----|
| Nexus base URL | `NEXUS_URL` |
| Maven repos | `NEXUS_MAVEN_RELEASES`, `NEXUS_MAVEN_SNAPSHOTS` |
| Auth | `NEXUS_ADMIN_USER`, `NEXUS_ADMIN_PASSWORD` |
| Jenkins cred ID (product) | `.coin/config.yaml` `jenkins.credentials` per project |

### D3: UI redirect

**Решение:** `/platform-settings` → redirect `/audit` (сохранить bookmark, не 404).

### D4: Admin nav

**Решение:** Admin group остаётся с единственным пунктом **Audit** (без пустой группы — Audit достаточно одного entry).

### D5: Seed scripts

**Решение:** удалить блок `api_put /v1/admin/platform/settings` из `seed-jenkins-lib-stack.sh`; bootstrap не зависит от PG settings.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Operator искал Nexus URL в UI | Runbook + docker README |
| External tooling вызывал API | BREAKING в OpenAPI changelog |
| Stale docs с platform_settings.runtime | Обновить golden-paths.md |

## Migration Plan

1. coin-api migration drop table + remove code paths.
2. coin-ui remove page + redirect.
3. Update seed script and docs.
4. `openspec validate --strict`.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| — | — | — | Нет blocking вопросов |
