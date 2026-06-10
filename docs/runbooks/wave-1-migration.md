# Wave 1 — миграция 50 repos (Control Plane v2)

> **⚠️ Corp gate:** rollout **заблокирован** — нет доступа в corp-сеть. До gate работаем только на **`samples/*`** (эталон: `demo-go-app`). Runbook — подготовка к corp.

**Ticket:** P1-06  
**Prerequisite:** P0 go/no-go **GO**, P1-01…P1-04 deployed.

## Цель

Перевести **50 продуктовых репозиториев** на config v2 + `Jenkinsfile.coin` + `coin-executor`.  
**Rollback на config v1 запрещён.**

## Scope pilot

| GP | Version | Статус |
|----|---------|--------|
| `go-app` | `1.0.0` | ✅ production-ready |

Другие GP (java/python) — Wave 2+ после publish в coin-api.

## Prerequisites (platform)

- [ ] coin-api `/ready`, auth policy согласована (`AUTH_DISABLED` только dev)
- [ ] Nexus: `coin-manifests`, `coin-executor/{ver}/`
- [ ] Jenkins credential `coin-api-token`
- [ ] `make endpoints` после restart docker/k3s
- [ ] coin-executor опубликован для целевых arch (arm64/amd64)

## Checklist на один repo

### 1. Inventory

| Поле | Значение |
|------|----------|
| Repo URL | |
| Текущий GP v1 | `template` / `templateVersion` |
| Целевой GP v2 | `go-app@1.0.0` |
| Jenkins multibranch | имя job |
| Owner / команда | |

### 2. Config v2

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
jenkins:
  credentials:
    docker: <credential-id>
project:
  name: <service-name>
  groupId: <team-domain>
  repository: Nexus_PROD
```

Mapping: [migrate-config-v1-to-v2.md](../how-to/migrate-config-v1-to-v2.md)

### 3. Jenkinsfile

Заменить v1 Shared Library pipeline (`coinPipeline()`) на  
[`coin-starters/Jenkinsfile.coin`](../../coin-starters/Jenkinsfile.coin).

Env overrides (optional):

| Var | Default |
|-----|---------|
| `COIN_API_URL` | `http://coin-api:8090` |
| `COIN_MANIFEST_CACHE_BASE` | Nexus `coin-manifests` |

### 4. Verify локально (optional)

```bash
curl -fsS -H "Authorization: Bearer $COIN_API_TOKEN" \
  "$COIN_API_URL/v1/golden-paths/go-app/versions/1.0.0/manifest" \
  -o .coin/manifest.json
coin-executor validate --project .coin/config.yaml --manifest .coin/manifest.json
```

### 5. Jenkins E2E

- [ ] Build **SUCCESS**: Resolve → Validate → Test → Build
- [ ] Stage **Report** → row в `build_reports` (coin-api DB)
- [ ] Docker image в registry

```sql
SELECT p.name, br.result, br.reported_at
FROM build_reports br JOIN projects p ON p.id = br.project_id
WHERE p.name = '<service-name>' ORDER BY br.id DESC LIMIT 1;
```

### 6. Sign-off

| | |
|---|---|
| Дата миграции | |
| Build URL | |
| Мигрировал | |
| Комментарий | |

## Rollout порядок

1. **Canary (5 repos)** — команды с высокой зрелостью CI, go-app
2. **Batch 1 (20 repos)** — после 5 green builds без инцидентов
3. **Batch 2 (25 repos)** — остаток до 50

Между batch: 24h мониторинг `coin_resolve_duration_seconds`, failed builds.

## Шаблон списка 50 repos

Заполнить PM / platform owner. Пример структуры:

| # | Repo | GP v2 | Batch | Status |
|---|------|-------|-------|--------|
| 1 | `coin/demo-go-app` | go-app@1.0.0 | canary | ✅ |
| 2 | | go-app@1.0.0 | canary | ☐ |
| … | | | | |
| 50 | | | batch-2 | ☐ |

> Полный список — в tracker команды (Jira/Linear). Эталон E2E: `samples/demo-go-app`.

## Escalation

| Проблема | Действие |
|----------|----------|
| Agent offline | `make endpoints`, см. [local-dev-control-plane.md](../how-to/local-dev-control-plane.md) |
| Manifest 404 | Проверить `goldenPath`/`version` в catalog |
| Report не пишется | coin-api logs, `POST /v1/builds/report` curl |
| GP не go-app | **не мигрировать** — ждать Wave 2 |

## Связанные документы

- [add-new-service-repo.md](../how-to/add-new-service-repo.md)
- [migrate-config-v1-to-v2.md](../how-to/migrate-config-v1-to-v2.md)
- [p0-go-no-go-checklist.md](../how-to/p0-go-no-go-checklist.md)
