## Context

Platform family **Runtime** (`/platform/runtime`) хранит `agent` components в PostgreSQL. Jenkins `publish-agent.sh` build → push → `POST .../versions/drafts` → **сейчас** `POST promote`.

Manifest resolve использует только:

```json
"runtime": { "image": "...", "digest": "sha256:..." }
```

`metadata.goarch` нигде не читается при resolve. Executor pin derived: `coin-agent@{ver}` → `coin-executor@{ver}` (hardcoded в `platform_runtime.go`).

Платформенные решения из explore (2026-06):

| Решение | Значение |
|---------|----------|
| Promote | Только руками на Platform |
| SoT pin | `image` + `digest` |
| Version | Semver = тег образа (bump в Jenkins) |
| GOARCH | Не хранить в platform metadata |
| Multi-profile | Да, не только `coin-agent` |

См. ADR [coin-ci-runtime](../../docs/adr/coin-ci-runtime.md).

## Goals / Non-Goals

**Goals:**

- Чёткий контракт agent version = registry record образа.
- CI path: register draft после push; human promote на UI.
- API gate: promote без digest → 422.
- UI: Path A (draft от CI) vs Path B (manual catch-up).
- Удалить `goarch` из agent metadata surface.
- Обобщить executor derive: убрать hardcoded switch по имени profile.
- Обновить ADR и runbooks.

**Non-goals:**

- Build-stacks / branching editor work (отдельный backlog).
- Executor independent lifecycle UI.
- Decoupled executor semver (agent@1.2 + executor@1.1) — не поддерживается; всегда same-V.

## Decisions

### D1. Agent metadata shape (v1 registry)

```yaml
# component_versions.metadata (agent only)
image: "nexus:8082/coin-docker/coin-agent:1.2.0"   # runtime ref в pod
digest: "sha256:abc..."                             # обязателен для promote
runtime: "coin-agent"                              # = components.name (profile)
```

**Удалить:** `goarch`.

**Инвариант promote:**

- `metadata.image` MUST match `*:{version}` (tag equals `component_versions.version`).
- `metadata.digest` MUST match `^sha256:[a-f0-9]{64}$`.

### D2. Promote gate — platform only

| Actor | Действие |
|-------|----------|
| `publish-agent.sh` | `POST drafts` only |
| Publisher (UI) | `POST .../promote` |

`publish-agent.sh` удаляет блок promote. Seed/E2E обновить: promote явным шагом в test или отдельным API call в e2e после UI simulation.

### D3. UI — два пути

```
Path A (CI)                         Path B (manual)
──────────                          ───────────────
CI → draft с image+digest           UI New draft:
UI release detail (draft)             Version (required)
  read-only image, digest             Image ref (required)
  [Promote]                           Digest (required)
                                      → save → Promote
```

`PlatformNewDraftPage` для agent: **не** optional metadata; image+digest required для осмысленного draft (или allow empty draft only if we want placeholder — **reject**: лучше требовать image+digest сразу для Path B).

GOARCH поля удалить из create и edit forms.

### D4. GOARCH — build-time only

`GOARCH` остаётся env в `publish-agent.sh` / `docker build --platform`. Не пишется в coin-api metadata. Digest достаточен для pin конкретного артефакта.

### D5. Executor derive — same version, любой agent profile

**Решение:** убрать hardcoded `switch` по `coin-agent` в `executorPinForAgentStack`. Для **любого** agent profile:

```
agent/{profile}@{V}  →  executor/coin-executor@{V}
```

**Runtime pod** (динамический Jenkins agent) использует только `metadata.image` + `metadata.digest` — этого достаточно для pull и старта. Секция `manifest.executor` собирается при resolve из PG (`executor/coin-executor@V`) для audit/pin; Jenkins **не** скачивает binary в bootstrap (baked в образе).

| Слой | Источник |
|------|----------|
| `manifest.runtime` | agent metadata (`image`, `digest`) |
| `manifest.executor` | derive same-V + metadata из `executor/coin-executor@V` в PG |

**Условие resolve:** `executor/coin-executor@{V}` MUST exist и быть visible (как сегодня для `coin-agent`). Публикуется отдельно (`publish-executor.sh`), не через форму runtime registry.

**Не в metadata agent:** `executorVersion`, `goarch`.

**Non-goal:** образы без baked `coin-executor` — не GP-eligible CI agents (out of scope).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| CI регистрирует draft, оператор забывает promote | UI badge «draft» на hub; GP pin требует published agent |
| Digest mismatch при ручном вводе | Формат validation; future: verify against registry |
| Breaking local scripts expecting auto-promote | Обновить docs + `make publish-agent` вывод «now promote in UI» |
| Executor@V отсутствует в PG при resolve | Как сегодня: fail resolve; publish-executor обязателен в CI pipeline |

## Migration Plan

1. coin-api: promote validation + generalize executor derive (можно deploy до UI).
2. publish-agent.sh: убрать promote.
3. coin-ui: формы + release detail.
4. Миграция данных: опционально SQL strip `goarch` из metadata JSON (не blocking).
5. Docs + E2E.

Rollback: вернуть auto-promote в script (нежелательно); API validation можно ослабить feature flag — не планируем.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Executor derive для non-`coin-agent` profiles | ✅ | Same-V для любого profile; убрать hardcoded switch (D5) |
| Q2 | Пустой draft без image (placeholder version) | ✅ | Отклонить — Path B требует image+digest при create |
