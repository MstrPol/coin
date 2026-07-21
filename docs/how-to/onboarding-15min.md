# Onboarding за 15 минут (Control Plane v2)

**Audience:** новый platform dev или инженер команды  
**Scope:** local docker стенд + green builds трёх build engines

## Что получите

- Рабочий стенд: Gitea, Nexus, Jenkins, k3s, coin-api, coin-ui
- Resolve manifest `go-app@*`
- Jenkins jobs **demo-go-app**, **demo-go-app-bp**, **demo-go-app-df** → SUCCESS
- coin-ui dashboard с projects registry

## Prerequisites (2 min)

- Docker Desktop запущен (8 GB+ RAM, 20+ GB disk для k3s)
- `make`, `curl` в PATH
- Клон integration workspace (sibling `coin-api`, `coin-executor`, … + `coin/`) — [workspace-layout](../workspace-layout.md)

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

## Шаг 2 — Platform + samples (5 min)

На Apple Silicon:

```bash
make publish-agent GOARCH=arm64   # coin-agent → Nexus
make coin-lib
make seed-jenkins-lib             # lib + gp-content + GP profiles
make samples                      # demo-go-app, demo-go-app-bp, demo-go-app-df
make coin-ui-up
```

coin-ui: http://localhost:8091 → Login → «Пропустить» (local) или `dev-local-admin-key`.

## Шаг 3 — Manifest resolve (1 min)

```bash
curl -sf "http://localhost:8090/v1/golden-paths/go-app/resolve?pin=*" \
  | jq '{gp: .goldenPath, engine: .build.engine, runtime: .runtime.image}'
```

## Шаг 4 — Jenkins E2E (4 min)

**Вариант A — один job:**

1. http://localhost:8080 → **demo-go-app** → main → **Build Now**
2. SUCCESS: Resolve → Bootstrap (podman) → Validate → Test → Build → Report

**Вариант B — все три build engines:**

```bash
make e2e-build-engines    # ~30 min, с prune k3s disk
```

| Job | GP | Engine |
|-----|-----|--------|
| demo-go-app | go-app | buildkit |
| demo-go-app-bp | go-app-bp | buildpack |
| demo-go-app-df | go-app-df | dockerfile |

## Шаг 5 — Optional (1 min)

coin-ui: **Projects** → `demo-go-app`.  
**GP Releases** → Detail → blast radius.

## Acceptance checklist

- [ ] `/ready` green
- [ ] manifest resolve OK (`build.engine` присутствует)
- [ ] demo-go-app build SUCCESS
- [ ] (optional) `make e2e-build-engines` 3/3
- [ ] coin-ui Dashboard показывает ≥1 project
- [ ] Документация: [docs/README.md](../README.md), [agent-build-model.md](../agent-build-model.md)

## Дальше по роли

| Роль | Документ |
|------|----------|
| Platform | [publish-gp-release.md](publish-gp-release.md), [agent-build-model.md](../agent-build-model.md) |
| Команда сервиса | [add-new-service-repo.md](add-new-service-repo.md) |
| PM | [fleet-analytics-pm.md](fleet-analytics-pm.md) |
| Миграция v1 | [migrate-config-v1-to-v2.md](migrate-config-v1-to-v2.md) |

## Troubleshooting

| Проблема | Решение |
|----------|---------|
| Agent offline | `make endpoints` |
| manifest 404 | `make seed-jenkins-lib` |
| Pod ephemeral-storage | `bash scripts/prune-k3s-disk.sh --all` |
| Старый bootstrap в логе | `make coin-lib` (очистка lib cache) |
| UI 401 | `AUTH_DISABLED=true` или admin key из `.env` |
