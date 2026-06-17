# Coin — локальный prod-like стенд

Docker Compose: Gitea, Nexus, k3s, Jenkins, coin-api, coin-ui. Platform-команды пушат артефакты в Gitea и регистрируют jobs в Jenkins.

| Сервис | Образ | Роль |
|--------|-------|------|
| **Gitea** | `gitea/gitea` | Git (platform + product repos) |
| **Nexus** | `sonatype/nexus3` | Maven + Docker registry `coin-docker` |
| **k3s** | `rancher/k3s` | Kubernetes для dynamic agents |
| **Jenkins** | custom LTS JDK17 | Platform CI + service pipelines |
| **coin-api** | build from monorepo | Resolve manifest, admin |
| **coin-ui** | build from monorepo | Admin SPA |

---

## Быстрый старт

```bash
cd docker
cp .env.example .env
make bootstrap
make endpoints

# Apple Silicon:
make publish-agent GOARCH=arm64
make coin-lib
make seed-jenkins-lib
make samples
make e2e-build-engines    # optional: 3 build engines E2E
make coin-ui-up
```

**Control Plane v2:** продуктовые repo — `@Library('coin-lib@1.0.0')` + `coinPipeline()`.

| URL | Логин |
|-----|-------|
| http://localhost:8080 | Jenkins — `admin` / см. `.env` |
| http://localhost:3000 | Gitea — `coin` / см. `.env` |
| http://localhost:8081 | Nexus — `admin` / см. `.env` |
| http://localhost:8090 | coin-api |
| http://localhost:8091 | coin-ui |
| https://localhost:8443 | k3s Dashboard — `make dashboard` |

Docker registry: `localhost:8082/coin-docker` (`coin` / `coin1234`).

Подробнее: [docs/how-to/local-dev-control-plane.md](../docs/how-to/local-dev-control-plane.md).

---

## Make-команды

### Инфраструктура

| Команда | Действие |
|---------|----------|
| `make bootstrap` | Первый подъём: build Jenkins, nexus, k3s, gitea, jenkins, postgres, coin-api |
| `make build-jenkins` | Пересобрать образ Jenkins (plugins) |
| `make endpoints` | Обновить k3s Endpoints (jenkins, nexus, gitea, coin-api) |
| `make up` / `make down` | Запуск / остановка |
| `make reset` | `down -v` — полный сброс данных |

### Platform → Gitea / Jenkins / Nexus

| Команда | Назначение |
|---------|------------|
| `make coin-executor` | Gitea `coin/coin-executor` + job |
| `make coin-gp-content` | Gitea `coin/coin-gp-content` + job |
| `make coin-lib` | Gitea tag `1.0.0` + Global Shared Library + job |
| `make publish-agent` | `coin-agent` image → Nexus + coin-api (`GOARCH=arm64` на Apple Silicon) |
| `make seed-jenkins-lib` | lib + gp-content + GP go-app / go-app-bp / go-app-df |
| `make coin-starters` | Gitea `coin/coin-starters` |
| `make samples` | demo repos → Gitea + multibranch jobs |
| `make e2e-mvp1` | Smoke resolve + Nexus без Jenkins |
| `make e2e-build-engines` | E2E: demo-go-app, demo-go-app-bp, demo-go-app-df |
| `make e2e-jenkins-lib` | API checks jenkins-lib model |

**Superseded:** `make coin-jenkins-agents`, `make agents-build`, `make coin-platform` (alias на удалённые agents).

### Demo-продукты

Манифест: [`samples.yaml`](samples.yaml). Клоны: `../samples/`.

| Repo | GP | Build engine |
|------|-----|--------------|
| demo-go-app | go-app | buildkit |
| demo-go-app-bp | go-app-bp | buildpack |
| demo-go-app-df | go-app-df | dockerfile |

```bash
make samples
```

---

## Gitea-репозитории

**Platform:** `coin/coin-executor`, `coin/coin-gp-content`, `coin/coin-lib`, `coin/coin-starters`

**Product demos:** `coin/demo-go-app`, `coin/demo-go-app-bp`, `coin/demo-go-app-df`, …

---

## Архитектура

```
bootstrap → Jenkins + k3s + Nexus + Gitea + coin-api
     │
     ├─ make publish-agent / seed-jenkins-lib  →  coin-agent + GP profiles
     └─ make samples                           →  product repos + multibranch

service pipeline (coinPipeline):
  resolve manifest → pod (coin-agent) → podman bootstrap → coin-executor stages
```

Pod-сеть k3s не видит docker-compose DNS — `make endpoints` регистрирует Service/Endpoints.

---

## Структура `docker/`

```
compose.yml
samples.yaml
Makefile
jenkins/                 # образ Jenkins, JCasC
k3s/                     # registries, dashboard
scripts/
  bootstrap.sh
  seed-jenkins-lib-stack.sh
  e2e-build-engines.sh
  prune-k3s-disk.sh
  publish-agent.sh       # → coin-executor/scripts/publish-agent.sh
  coin-lib.sh
  samples.sh
  ...
```

---

## Переменные окружения

Файл `.env` (из `.env.example`).

| Переменная | Default | Назначение |
|------------|---------|------------|
| `DOCKER_PLATFORM` | `linux/arm64` | платформа образов |
| `JENKINS_HTTP_PORT` | `8080` | Jenkins UI |
| `GITEA_HTTP_PORT` | `3000` | Gitea |
| `NEXUS_HTTP_PORT` | `8081` | Nexus UI |
| `NEXUS_DOCKER_PORT` | `8082` | Docker registry |

---

## Troubleshooting

**Bootstrap долго** — 5–15 мин при первой сборке Jenkins.

**OOM (exit 137)** — 8 GB+ RAM Docker.

**Agent offline** — `make endpoints`.

**k3s disk full (buildpack)** — `bash scripts/prune-k3s-disk.sh --all`.

**Stale Jenkins lib** — `make coin-lib`.

**Gitea clone failed** — `make coin-lib seed-jenkins-lib samples`.

---

## Связанные документы

- [coin-starters/README.md](../coin-starters/README.md)
- [docs/jenkins-setup.md](../docs/jenkins-setup.md)
- [docs/agent-build-model.md](../docs/agent-build-model.md)
- [docs/architecture.md](../docs/architecture.md)
