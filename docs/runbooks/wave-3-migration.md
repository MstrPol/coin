# Wave 3 — full fleet 1500+ repos

> **⚠️ Corp gate:** rollout **заблокирован** до доступа в corp-сеть и завершения P2–P3 на prod. Документ — подготовка runbook + comms templates (P3-05).

**Ticket:** P3-04 (runbook), P3-05 (comms)  
**Prerequisite:** Wave 1 (50) ✅, Wave 2 (500) ✅, scanner CronJob green, coin-ui analytics.

## Цель

Перевести **оставшийся fleet (~1500+ repos)** на Control Plane v2 с полным registry coverage и adoption tracking.

## Scope

| GP | Минимум | Примечание |
|----|---------|------------|
| `go-app` | `1.0.0` | bulk Wave 1–2 |
| `java-app`, `python-app` | TBD | после GP publish в coin-api |
| Legacy v1 config | — | scanner skip → migration mandatory |

**Rollback на config v1 запрещён.**

## Prerequisites (platform)

- [ ] Fleet scanner CronJob: last success < 24h ([scanner-ops.md](scanner-ops.md))
- [ ] coin-ui доступен PM (SSO — P4-02 optional)
- [ ] `catalog_policy.minimum` согласован с PM
- [ ] Blast radius проверен для target GP release
- [ ] Nexus manifest cache warm для новых GP versions
- [ ] On-call: [api-down-nexus-fallback.md](api-down-nexus-fallback.md)

## Rollout порядок

1. **Inventory sync** — scanner full rescan (`force=true`), сверка с tracker
2. **Stale comms** — projects > 90d без build ([SQL](../how-to/fleet-analytics-pm.md#stale-projects-нет-build-90-дней))
3. **Batch migration** — 100–200 repos/week (rate limit platform + Jenkins)
4. **GP bump waves** — после majority на baseline version
5. **Minimum enforce** — только когда < 5% на EOL version

Между batch: мониторинг `coin_scan_repos_failed`, failed builds, resolve 403.

## Tracker (PM)

| # | Repo | GP target | Batch | Owner | Status | Last seen |
|---|------|-----------|-------|-------|--------|-----------|
| 1 | | go-app@1.0.0 | w3-b1 | | ☐ | |
| … | | | | | | |

Источник **Last seen**: coin-ui Projects или SQL `project_bindings.last_seen_at`.

---

## Comms templates

### 1. Wave 3 kickoff (all engineering)

**Subject:** Coin Control Plane v2 — Wave 3 migration (action required)

**Body:**

> Команда,
>
> Завершаем переход fleet на Coin Control Plane v2 (config v2 + coin-executor).
>
> **Что нужно от вас:**
> 1. Обновить `.coin/config.yaml` — см. [migrate-config-v1-to-v2](../how-to/migrate-config-v1-to-v2.md)
> 2. Заменить Jenkinsfile на `Jenkinsfile.coin`
> 3. Прогнать green build в Jenkins
>
> **Дедлайн для batch {BATCH_ID}:** {DATE}  
> **Ваши repos:** {REPO_LIST_OR_TRACKER_LINK}
>
> **Поддержка:** #coin-platform, office hours {SCHEDULE}
>
> Rollback на config v1 **не поддерживается**.

### 2. GP version bump (adoption)

**Subject:** [Coin] Upgrade golden path {GP_NAME} {OLD_VER} → {NEW_VER}

**Body:**

> Platform опубликовала `{GP_NAME}@{NEW_VER}`.
>
> **Blast radius:** {N} repos на `{OLD_VER}` ([dashboard link](http://coin-ui/releases/{GP_NAME}/{OLD_VER})).
>
> **Рекомендуемое окно:** {DATE_RANGE}  
> **Изменения:** {RELEASE_NOTES_LINK}
>
> Обновите в `.coin/config.yaml`:
> ```yaml
> coin:
>   goldenPath: {GP_NAME}
>   version: "{NEW_VER}"
> ```
>
> После {MINIMUM_ENFORCE_DATE} builds на `{OLD_VER}` будут **заблокированы** (catalog minimum).

### 3. Stale repo / no CI activity

**Subject:** [Coin] Repo {REPO_NAME} — stale CI ({DAYS} days)

**Body:**

> Repo `{REPO_NAME}` не имеет успешного build / scanner activity **{DAYS}** дней.
>
> Текущий GP binding: `{GP_NAME}@{VERSION}` (если известен).
>
> Пожалуйста, подтвердите одно из:
> - [ ] Repo active — запустите Jenkins build или обновите config v2
> - [ ] Repo archived — ответьте для исключения из fleet tracker
> - [ ] Нужна помощь с миграцией — ticket в #coin-platform
>
> **Deadline ответа:** {DATE}

### 4. EOL / minimum version enforcement

**Subject:** [Coin] ACTION REQUIRED — {GP_NAME} {EOL_VER} end-of-life {DATE}

**Body:**

> С {DATE} `catalog_policy.minimum` для `{GP_NAME}` = `{MIN_VER}`.
>
> Resolve manifest для `{EOL_VER}` вернёт **403** — builds не стартуют.
>
> **Затронуто repos:** {N} — список: {TRACKER_FILTER_LINK}
>
> Миграция: обновить `coin.version` в config + green build **до {DATE}**.

### 5. Migration complete (batch sign-off)

**Subject:** [Coin] Wave 3 batch {BATCH_ID} complete ✅

**Body:**

> Batch **{BATCH_ID}** ({N} repos) мигрирован на v2.
>
> **Metrics:**
> - Projects in registry: {TOTAL}
> - On `{GP}@{VER}`: {COUNT}
> - Failed builds (7d): {FAIL_COUNT}
>
> Следующий batch: **{NEXT_BATCH_ID}** старт {DATE}.

---

## Escalation

| Проблема | Runbook |
|----------|---------|
| Scanner down | [scanner-ops.md](scanner-ops.md) |
| API down | [api-down-nexus-fallback.md](api-down-nexus-fallback.md) |
| Resolve 403 после minimum bump | PM comms template #4 + owner migration |
| Repo не в registry | v1 config / нет `.coin/config.yaml` |

## Связанные документы

- [wave-migration-checklist.md](../how-to/wave-migration-checklist.md)
- [wave-1-migration.md](wave-1-migration.md)
- [fleet-analytics-pm.md](../how-to/fleet-analytics-pm.md)
- [scanner-ops.md](scanner-ops.md)
