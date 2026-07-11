# Scanner ops — fleet scan runbook

> **SUPERSEDED (UI-02, 2026-06):** Fleet scanner удалён. Projects registry — только через `POST /v1/builds/report`. CronJob и `make scan-fleet` больше не используются.

**Ticket:** P3-05  
**Audience:** Platform / DevOps on-call  
**Component:** ~~`coin-api/cmd/scanner`~~, ~~CronJob `coin-fleet-scanner`~~

## Назначение

Fleet scanner обходит repos пользователя `GITEA_ORG` в Gitea, читает `.coin/config.yaml` v2 и upsert'ит:

- `projects`
- `project_bindings` (gp_name, gp_version, git_url, last_seen_at)
- `scanner_state` (incremental SHA cache)

**Не делает:** clone repo, build, publish. v1 config (`template`/`templateVersion`) → skip.

## Архитектура

```mermaid
flowchart LR
  subgraph trigger [Trigger]
    CronJob[CronJob 02:00]
    CLI[make scan-fleet]
    API[POST /v1/admin/scan]
  end
  subgraph scan [Scanner]
    Gitea[Gitea API]
    PG[(PostgreSQL)]
  end
  subgraph metrics [Observability]
    Prom[/metrics on coin-api]
  end
  CronJob --> API
  CLI --> scan
  API --> scan
  scan --> Gitea
  scan --> PG
  API --> Prom
```

| Способ запуска | Когда использовать |
|----------------|-------------------|
| **CronJob** (prod/local k3s) | Nightly incremental scan |
| **POST /v1/admin/scan** | Manual trigger + Prometheus metrics |
| **CLI** `make scan-fleet` | Local dev с host → localhost Gitea/Postgres |

**SLA (target):** ~1500 repos < **2h** (incremental; full rescan реже).

## Env

| Variable | Default (local) | Description |
|----------|-----------------|-------------|
| `DATABASE_URL` | required | PostgreSQL DSN |
| `GITEA_URL` | `http://gitea:3000` | Gitea base URL |
| `GITEA_USER` / `GITEA_PASSWORD` | `coin` | API auth |
| `GITEA_ORG` | `coin` | User/org — список repos |

CronJob использует coin-api endpoint → env scanner наследует coin-api container.

## Local operations

### CLI (host)

```bash
cd docker
make scan-fleet              # incremental
# force rescan:
cd ../coin-api && go run ./cmd/scanner -force
```

### API (metrics update)

```bash
curl -X POST http://localhost:8090/v1/admin/scan \
  -H "X-API-Key: dev-local-admin-key"

curl -sf http://localhost:8090/metrics | grep '^coin_scan'
```

### K8s CronJob (docker-compose k3s)

```bash
cd docker
make endpoints               # coin-api Endpoints в k3s
make scan-cronjob-apply      # apply CronJob + Secret

# Ручной Job (не ждать 02:00):
make scan-cronjob-run
```

Manifests: `docker/k3s/coin-scanner-cronjob.yaml`, `coin-scanner-job.yaml`.

**Secret `coin-admin`:** ключ `api-key` = `COIN_ADMIN_API_KEY`. В prod — заменить на vault/ExternalSecret.

## Prometheus metrics

| Metric | Type | Meaning |
|--------|------|---------|
| `coin_scan_duration_seconds` | histogram | Длительность последнего scan |
| `coin_repos_scanned` | gauge | Repos обновлены в этом run |
| `coin_scan_repos_total` | gauge | Всего repos в org |
| `coin_scan_repos_skipped` | gauge | Skip (unchanged SHA, no config, v1) |
| `coin_scan_repos_failed` | gauge | Ошибки fetch/parse/upsert |
| `coin_scan_last_success_timestamp` | gauge | Unix time последнего успешного scan |

**Alerting (рекомендация prod):**

- `coin_scan_last_success_timestamp` stale > 26h → page platform
- `coin_scan_repos_failed` > 0 → warning, investigate logs
- `coin_scan_duration_seconds` p99 > 7200 → SLA breach

## Incremental logic

1. List repos org `GITEA_ORG`
2. SHA default branch (`GET /repos/{owner}/{repo}/branches/{branch}`)
3. Compare с `scanner_state.last_sha` — если equal → **skip**
4. Fetch `.coin/config.yaml` raw
5. Parse v2 (`coin.goldenPath`, `coin.version`, `project.name`)
6. Upsert project + binding; save SHA

**Force rescan:** `-force` / `?force=true` — игнорирует SHA cache.

## Exit codes / HTTP status

| Outcome | CLI exit | API HTTP |
|---------|----------|----------|
| Success, 0 failed | 0 | 200 |
| Success, N failed repos | 1 | 207 Multi-Status |
| Fatal (DB, Gitea list) | 1 | 500 |

Failed repos **не блокируют** остальные — счётчик `reposFailed`, warn в logs.

## Troubleshooting

### CronJob Failed / CrashLoop

```bash
docker compose exec -T k3s kubectl get cronjob coin-fleet-scanner
docker compose exec -T k3s kubectl get jobs -l app=coin-fleet-scanner
docker compose exec -T k3s kubectl logs job/coin-fleet-scanner-manual
```

| Symptom | Cause | Fix |
|---------|-------|-----|
| `connection refused coin-api:8090` | Endpoints не зарегистрированы | `make endpoints` |
| `401 Unauthorized` | Secret key mismatch | Sync `coin-admin` Secret с `COIN_ADMIN_API_KEY` |
| Job timeout 7200s | Fleet слишком большой / Gitea slow | Parallel workers (roadmap), increase `activeDeadlineSeconds` |
| curl exit 22 | API 500 | coin-api logs |

### Gitea errors

```bash
docker compose logs coin-api | grep -i gitea
# Test from coin-api container:
docker compose exec coin-api wget -qO- http://gitea:3000/api/v1/version
```

| Log | Fix |
|-----|-----|
| `branch sha` failed | Repo empty / no default branch |
| `fetch config` 404 | No `.coin/config.yaml` — expected skip after mark |
| `skip v1 config` | Owner must migrate to v2 |

### DB / migrations

Scanner CLI и API run migrations on start (CLI only). coin-api уже применил `006_scanner_state.sql`.

```sql
SELECT COUNT(*) FROM scanner_state;
SELECT repo_full_name, last_scan_at FROM scanner_state ORDER BY last_scan_at DESC LIMIT 10;
```

### Incremental «ничего не сканирует»

Expected если SHA не изменился. Verify:

```bash
curl -X POST 'http://localhost:8090/v1/admin/scan?force=true' \
  -H 'X-API-Key: dev-local-admin-key'
```

## Verify scan result

```sql
SELECT COUNT(DISTINCT p.name) FROM projects p;
SELECT pb.gp_name, pb.gp_version, COUNT(*) 
FROM project_bindings pb
GROUP BY 1, 2 ORDER BY 1, 2;
```

coin-ui: **Projects** page или **Dashboard** counters.

## Corp rollout checklist

- [ ] CronJob в prod k8s namespace control-plane
- [ ] `COIN_ADMIN_API_KEY` из secret manager
- [ ] Gitea SA read-only на product org(s)
- [ ] Prometheus alerts на metrics выше
- [ ] Runbook linked in on-call playbook
- [ ] PM access: [fleet-analytics-pm.md](../how-to/fleet-analytics-pm.md)

## Связанные документы

- [coin-api/README.md](../../coin-api/README.md) — scanner CLI, env
- [fleet-analytics-pm.md](../how-to/fleet-analytics-pm.md) — PM view
- [local-dev-control-plane.md](../how-to/local-dev-control-plane.md)
- [docker/README.md](../../docker/README.md)
