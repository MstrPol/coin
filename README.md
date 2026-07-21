# Coin — Control Plane v2

```
 ██████╗ ██████╗ ██╗███╗   ██╗
██╔════╝██╔═══██╗██║████╗  ██║
██║     ██║   ██║██║██╔██╗ ██║
██║     ██║   ██║██║██║╚██╗██║
╚██████╗╚██████╔╝██║██║ ╚████║
 ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝
```

Платформа CI: продукт указывает `coin.goldenPath` + `coin.version`; платформа выпускает GP releases и resolved manifest.

**Layout:** sibling-репозитории + meta `coin/` — [docs/workspace-layout.md](docs/workspace-layout.md).  
Corp extract — [docs/runbooks/prod-repo-split.md](docs/runbooks/prod-repo-split.md) (после corp gate).

## Где что лежит

| Каталог (workspace) | Назначение |
|---------------------|------------|
| `../coin-api/` | Resolve, Admin API, seed |
| `../coin-executor/` | CLI + `coin-agent` image |
| `../coin-lib/` | Jenkins Shared Library |
| `../coin-ui/` | Platform SPA |
| [`coin-starters/`](coin-starters/README.md) | Product scaffolding |
| [`docker/`](docker/README.md) | Local compose стенд |
| [`docs/`](docs/README.md) | Документация |
| [`openspec/`](openspec/) | Канон требований |
| [`samples/`](samples/) | E2E demos |

## Onboarding

→ [docs/how-to/onboarding-15min.md](docs/how-to/onboarding-15min.md)

```bash
cd docker
cp .env.example .env
make bootstrap && make endpoints
make publish-agent GOARCH=arm64   # Apple Silicon; иначе omit GOARCH
make coin-lib
make seed-jenkins-lib
make coin-starters && make samples
make coin-ui-up
curl -sf http://localhost:8090/ready
```

## Продуктовый контракт

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
```

Jenkinsfile — [`coin-starters/Jenkinsfile.coin`](coin-starters/Jenkinsfile.coin).  
GP composition: **2 pin** (`agent`, `branching-model`) + embedded pipeline.

## Документация

Индекс — [docs/README.md](docs/README.md). OpenSpec — [openspec/specs/](openspec/specs/).
