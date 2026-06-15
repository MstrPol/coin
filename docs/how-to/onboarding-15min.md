# Onboarding за 15 минут (Control Plane v2)

**Ticket:** P4-04  
**Audience:** новый platform dev или инженер команды  
**Scope:** local docker стенд + один green build `demo-go-app`

## Что получите

- Рабочий стенд: Gitea, Nexus, Jenkins, k3s, coin-api, coin-ui
- Resolve manifest `go-app@1.0.0`
- Jenkins build **demo-go-app** SUCCESS
- coin-ui dashboard с projects registry

## Prerequisites (2 min)

- Docker Desktop запущен (8 GB+ RAM)
- `make`, `curl` в PATH
- Клон monorepo `coin/`

## Шаг 1 — Infra (5 min)

```bash
cd docker
cp .env.example .env    # если нет
make bootstrap
make endpoints          # обязательно после bootstrap/restart
```

Verify:

```bash
curl -sf http://localhost:8090/ready | jq .
# → {"status":"ready"}
```

| URL | Назначение |
|-----|------------|
| http://localhost:8080 | Jenkins (`admin` / см. `.env`) |
| http://localhost:8090 | coin-api |
| http://localhost:8091 | coin-ui |
| http://localhost:3000 | Gitea |

## Шаг 2 — Platform + samples (3 min)

```bash
make coin-jenkins-agents   # agents + catalog → Gitea
make samples            # demo-go-app → Gitea + Jenkins job
make coin-ui-up
```

coin-ui: http://localhost:8091 → Login → «Пропустить» (local) или `dev-local-admin-key`.

## Шаг 3 — Manifest resolve (1 min)

```bash
curl -sf http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest \
  | jq '{gp: .goldenPath, runtime: .runtime.image, executor: .executor.version}'
```

Nexus cache (fallback path — pointer → blob):

```bash
SNAPSHOTS=http://localhost:8081/repository/maven-snapshots
curl -sf "${SNAPSHOTS}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.0.json" | jq '{manifestHash, blobUrl}'
```

## Шаг 4 — Jenkins E2E (4 min)

1. Открыть http://localhost:8080
2. Job **demo-go-app** → branch **main** → **Build Now**
3. Дождаться **SUCCESS** (Resolve → Validate → Test → Build → Report)

Проверка report в coin-ui: **Projects** → `demo-go-app` на `go-app@1.0.0`.

## Шаг 5 — Optional (1 min)

```bash
# Projects registry — автоматически при первом build report (fleet scanner удалён)
curl -sf http://localhost:8090/metrics | grep coin_scan
```

coin-ui: **GP Releases** → Detail → blast radius chart.

## Acceptance checklist

- [ ] `/ready` green
- [ ] manifest resolve OK
- [ ] demo-go-app build SUCCESS
- [ ] coin-ui Dashboard показывает ≥1 project
- [ ] Понятно где docs: [docs/README.md](../README.md)

## Дальше по роли

| Роль | Документ |
|------|----------|
| Platform | [publish-gp-release.md](publish-gp-release.md), [coin-ui-user-guide.md](../coin-ui-user-guide.md) |
| Команда сервиса | [add-new-service-repo.md](add-new-service-repo.md) |
| PM | [fleet-analytics-pm.md](fleet-analytics-pm.md) |
| Миграция v1 | [migrate-config-v1-to-v2.md](migrate-config-v1-to-v2.md) |

## Troubleshooting

| Проблема | Решение |
|----------|---------|
| Agent offline | `make endpoints` |
| manifest 404 | `make coin-jenkins-agents`, проверить GP seed в postgres |
| executor 404 | Опубликовать binary в Nexus — см. [local-dev-control-plane.md](local-dev-control-plane.md) |
| UI 401 | `AUTH_DISABLED=true` или admin key из `.env` |
