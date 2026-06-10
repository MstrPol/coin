# Fleet analytics для PM

**Ticket:** P3-05  
**Audience:** PM, platform owner, release manager  
**Scope:** чтение adoption и blast radius **без curl** — через coin-ui и SQL (corp gate для full fleet).

## Зачем

При fleet **1500+ repos** вручную grep по Gitea не масштабируется. Control plane собирает registry из:

| Источник | Что пишет | Когда |
|----------|-----------|-------|
| **Fleet scanner** | `projects`, `project_bindings`, `git_url` | Nightly CronJob / `make scan-fleet` |
| **Build report** | `build_reports`, обновляет `last_seen_at` | Каждый успешный Jenkins build |

PM использует analytics **до** publish GP bump — чтобы понять blast radius и план коммуникации.

## coin-ui (рекомендуемый путь)

**URL (local):** http://localhost:8091  
**Admin key:** `COIN_ADMIN_API_KEY` (local: `dev-local-admin-key`)

Подробнее: [coin-ui-user-guide.md](../coin-ui-user-guide.md).

### Dashboard

Счётчики fleet:

- **Projects** — уникальные сервисы в registry
- **GP releases** — опубликованные версии golden path
- **Build reports** — builds, отправившие report в coin-api
- **Golden paths** — distinct GP names

Клик по карточке → Projects или GP Releases.

### Projects — «кто на какой версии»

Таблица с **последним GP binding** на project.

Фильтры (URL или форма):

- `goldenPath=go-app`
- `version=1.0.0`

Пример: «сколько projects на go-app 1.0.0» → фильтр обоих полей, число строк в таблице.

Колонка **Last seen** — последняя активность (scanner или build report).

### GP Releases → Detail → Blast radius

1. **GP Releases** — список опубликованных releases
2. **Detail** на строке → страница release (composition, manifest hash, git tag)
3. **Blast radius chart** — bar chart по версиям GP:
   - **На этой версии** — projects, которые сейчас на target version
   - **На других версиях** — остальные на том же GP
   - **На более старых** — semver ниже target (кандидаты на bump)

Ссылка **«Показать projects →»** открывает Projects с фильтром по версии.

> **Acceptance (P3-03):** PM видит «847 repos on go-app 1.0.0» без API curl — через Detail + chart или Projects filter.

### Components (read-only)

Registry platform-компонентов (executor, agent, pipeline, validate, dockerfile):

- latest version
- count published versions
- дата последнего publish

Publish — только platform через Admin API ([publish-gp-release.md](publish-gp-release.md)).

## Blast radius перед GP publish

**Workflow PM + platform:**

1. Platform готовит GP release `go-app@1.0.4`
2. PM открывает **GP Releases → go-app → 1.0.4 (draft/preview)** или текущий latest
3. Смотрит blast radius для **текущей** production version (например `1.0.0`):
   - сколько projects на `1.0.0`
   - сколько на older (`0.x`) — приоритет для comms
4. Platform publish → PM мониторит adoption через Projects filter `version=1.0.4`

API (если нужен скрипт):

```bash
curl -sf -H "X-API-Key: ${COIN_ADMIN_API_KEY}" \
  "http://coin-api:8090/v1/admin/golden-paths/go-app/versions/1.0.0/blast-radius" | jq .
```

## Stale projects (нет build 90+ дней)

> **Pilot:** метрика `stale_projects` в dashboard — roadmap. Сейчас — SQL + Projects **Last seen**.

**Определение stale:** project без build report и без scanner update **> 90 дней**.

```sql
-- Projects stale > 90d (последний binding)
SELECT p.name,
       pb.gp_name,
       pb.gp_version,
       pb.last_seen_at,
       pb.git_url
FROM projects p
JOIN LATERAL (
    SELECT gp_name, gp_version, git_url, last_seen_at
    FROM project_bindings pb2
    WHERE pb2.project_id = p.id
    ORDER BY pb2.last_seen_at DESC
    LIMIT 1
) pb ON true
WHERE pb.last_seen_at < now() - interval '90 days'
ORDER BY pb.last_seen_at ASC;
```

**Действия PM:**

| Ситуация | Действие |
|----------|----------|
| Repo archived / decommissioned | Убрать из tracker, не включать в wave |
| Repo active, но нет CI | Escalation owner — включить multibranch |
| На старой GP version | Wave comms + migration slot |
| Нет `.coin/config.yaml` v2 | Scanner skip — owner мигрирует config |

## Minimum version и deprecated GP

При resolve coin-api enforce `catalog_policy.minimum`:

- версия **ниже minimum** → **403 Forbidden** (build не стартует на EOL GP)
- версия в `deprecated[]` → **200 + Warning** header

PM следит: после bump `minimum` — projects на старых версиях **не смогут resolve** → нужен migration wave **до** или **сразу после** publish.

## Типовые вопросы PM

| Вопрос | Где смотреть |
|--------|--------------|
| Сколько repos на go-app 1.0.0? | Projects filter или GP Detail blast chart |
| Кто ещё не на v2 config? | Scanner skip v1 — нет row в Projects; см. Gitea inventory |
| Сколько repos затронет bump 1.0.0 → 1.0.1? | Blast radius **onThisVersion** для 1.0.0 |
| Когда последний scan? | Prometheus `coin_scan_last_success_timestamp` или [scanner-ops.md](../runbooks/scanner-ops.md) |
| Какой composition у release? | GP Releases → Detail |

## Ограничения (pilot / corp gate)

| Доступно сейчас (local) | После corp gate |
|-------------------------|-----------------|
| `samples/*` + scanner на local Gitea | Full corp Gitea fleet (~1500 repos) |
| coin-ui read-only | SSO (P4-02), publish wizard (P4-01) |
| SQL stale queries | Dashboard metric `stale_projects` (roadmap) |

## Связанные документы

- [coin-ui-user-guide.md](../coin-ui-user-guide.md)
- [publish-gp-release.md](publish-gp-release.md)
- [scanner-ops.md](../runbooks/scanner-ops.md)
- [wave-3-migration.md](../runbooks/wave-3-migration.md) — Wave 3 + comms templates
- [wave-migration-checklist.md](wave-migration-checklist.md)
