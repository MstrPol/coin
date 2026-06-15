# Coin — Control Plane v2

```
 ██████╗ ██████╗ ██╗███╗   ██╗
██╔════╝██╔═══██╗██║████╗  ██║
██║     ██║   ██║██║██╔██╗ ██║
██║     ██║   ██║██║██║╚██╗██║
╚██████╗╚██████╔╝██║██║ ╚████║
 ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝
```

Платформа CI для **1500+ сервисов**: продукт указывает `coin.goldenPath` + `coin.version`, платформа централизованно выпускает golden path releases и manifest.

**Monorepo (dev):** все компоненты v2 в одном репозитории до corp split ([P4-03 runbook](docs/runbooks/prod-repo-split.md)).

## Компоненты v2

| Каталог | Назначение |
|---------|------------|
| [`coin-api/`](coin-api/README.md) | Resolve manifest, registry, admin API |
| [`coin-executor/`](coin-executor/README.md) | Runtime pipeline: validate, stages, build report |
| [`coin-ui/`](coin-ui/README.md) | Admin SPA: dashboard, publish wizard, audit log |
| [`coin-jenkins-agents/`](coin-jenkins-agents/README.md) | CI agent images (`agents-build`) |
| [`coin-starters/`](coin-starters/README.md) | Скелетоны product repos + thin Jenkinsfile |
| [`docker/`](docker/README.md) | Local prod-like стенд (Gitea, Nexus, k3s, Jenkins) |
| [`samples/`](samples/demo-go-app/README.md) | E2E эталон (`demo-go-app`) |

## Onboarding за 15 минут

→ [docs/how-to/onboarding-15min.md](docs/how-to/onboarding-15min.md)

Кратко для platform dev:

```bash
cd docker
cp .env.example .env
make bootstrap && make endpoints
make coin-jenkins-agents && make coin-starters && make samples
make coin-ui-up          # http://localhost:8091
curl -sf http://localhost:8090/ready
```

Jenkins: http://localhost:8080 → **demo-go-app** → Build Now → SUCCESS.

## Продуктовый контракт

```yaml
# .coin/config.yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
```

Jenkinsfile — [`coin-starters/Jenkinsfile.coin`](coin-starters/Jenkinsfile.coin).

## Документация

Полный индекс — [docs/README.md](docs/README.md).

| Роль | Старт |
|------|-------|
| Platform dev | [local-dev-control-plane](docs/how-to/local-dev-control-plane.md) |
| Команда сервиса | [add-new-service-repo](docs/how-to/add-new-service-repo.md) |
| PM / analytics | [coin-ui-user-guide](docs/coin-ui-user-guide.md) |
| Миграция v1→v2 | [migrate-config-v1-to-v2](docs/how-to/migrate-config-v1-to-v2.md) |

## Структура monorepo

```
coin/
├── coin-api/           # HTTP API (control plane)
├── coin-executor/      # CLI runtime
├── coin-ui/            # Admin SPA
├── coin-jenkins-agents/  # CI agent stacks + catalog
├── coin-starters/        # Product repo scaffolding
├── docker/             # Compose стенд
├── docs/
├── samples/            # demo-go-app, …
```
