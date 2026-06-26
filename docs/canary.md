# Canary rollout

Platform-first canary для pin `*` — audience rules через coin-api (не Nexus fallback).

## Модель

| Catalog field | Назначение |
|---------------|------------|
| `latest` | Stable line — Nexus pointer `latest.json`, pin `*` без project |
| `latest_canary` | Canary line — только через API resolve с `project` |

**GP-level canary:** `latest_canary` может указывать на GP release со `status=draft`. Это намеренно нестабильная линия для pilot-проектов.

**Component lifecycle:** у платформенных компонентов только `draft` → `published`. Статуса `canary` на уровне компонентов нет.

**Draft pins на canary line:** при resolve canary для `gp-content` / `branching-model` допускаются pins со `status=draft` (PG content_ref). Agent pins — только `published`.

## Precedence (resolve)

При `GET /v1/golden-paths/{name}/resolve?pin=*&project=…`:

1. `canary_mode=canary` на project → **always canary**
2. `canary_mode=stable` → **always stable**
3. `canary_mode=default` → `hash(project) % 100 < canary_percent` → canary, иначе stable

Без `project` при pin `*` → **stable** (safe default).

Response header: `X-Coin-Channel: stable|canary`.

## GP promote gate

Promote GP `draft` → `published` блокируется, если любой pin в composition не `published`. API: HTTP 409 + `blockingPins[]`.

## Fallback (API down)

Nexus pointer `latest.json` = **stable only**. Canary line доступна только через coin-api.

## Health (build reports)

Executor отправляет `channel`, `requestedPin`, `failedStage` в `POST /v1/builds/report`.

Admin: `GET /v1/admin/golden-paths/{name}/versions/{version}/health?channel=canary`

Пороги в `canary_policy`: degraded / critical failure rate. Auto-rollback **не** реализован — только signal в coin-ui.

## coin-ui

- **Canary** — slider %, preview resolve, health badge; warning при GP draft на canary line
- **Projects** — canary mode per project
- **Catalog** — `latest_canary` version
- **GP release** — draft pin warnings, promote disabled до publish всех pins
