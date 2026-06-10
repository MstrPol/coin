# Coin — локальный prod-like стенд

Docker Compose: Gitea, Nexus, k3s, Jenkins. Platform-команды пушат артефакты в Gitea и регистрируют jobs в Jenkins.

| Сервис | Образ | Роль |
|--------|-------|------|
| **Gitea** | `gitea/gitea` | Git (platform + product repos) |
| **Nexus** | `sonatype/nexus3` | Maven (из коробки) + Docker registry `coin-docker` |
| **k3s** | `rancher/k3s` | Kubernetes для dynamic agents |
| **Jenkins** | custom LTS JDK17 | Platform CI + service pipelines |

---

## Быстрый старт

```bash
cd docker
cp .env.example .env          # или make bootstrap создаст сам
make bootstrap                # infra: nexus, k3s, gitea, jenkins, postgres, coin-api
make endpoints                # после restart compose — обновить k3s Endpoints (jenkins/nexus/gitea)

make coin-jenkins-agents            # agents → Gitea coin/coin-jenkins-agents
make coin-starters                  # starters → Gitea (optional)
make coin-platform                  # alias: agents + starters
make coin-executor            # executor → Gitea + job coin-executor
make agents-build             # job agents-build (JCasC)
make samples                  # demo-продукты → samples/ + Gitea + multibranch jobs
make coin-api-up              # только postgres + coin-api
make coin-ui-up               # coin-api + coin-ui → http://localhost:8091
make scan-fleet               # Gitea scanner → projects registry (P3-01)
make scan-cronjob-apply       # K8s CronJob nightly scan (P3-02)
make scan-cronjob-run         # one-off scan Job в k3s
```

**Control Plane v2:** продуктовые repo используют `Jenkinsfile.coin` + `coin-executor`.

| URL | Логин |
|-----|-------|
| http://localhost:8080 | Jenkins — `admin` / см. `.env` |
| http://localhost:3000 | Gitea — `coin` / см. `.env` |
| http://localhost:8081 | Nexus — `admin` / см. `.env` |
| https://localhost:8443 | k3s Dashboard — `make dashboard` |

Docker registry: `localhost:8082/coin-docker` (`coin` / `coin1234`).

Подробнее: [docs/how-to/local-dev-control-plane.md](../docs/how-to/local-dev-control-plane.md).

---

## Make-команды

### Инфраструктура

| Команда | Действие |
|---------|----------|
| `make bootstrap` | Первый подъём: build Jenkins (если нет образа), nexus, k3s, gitea, jenkins |
| `make build-jenkins` | Пересобрать образ Jenkins (plugins) |
| `make up` / `make down` | Запуск / остановка (volumes сохраняются) |
| `make reset` | `down -v` — полный сброс данных |
| `make dashboard` | Kubernetes Dashboard (опционально) |

### Platform → Gitea / Jenkins

| Команда | Gitea | Jenkins |
|---------|-------|---------|
| `make coin-jenkins-agents` | `coin/coin-jenkins-agents` | sync `catalog.yaml` из Gitea |
| `make coin-starters` | `coin/coin-starters` | — |
| `make coin-platform` | оба repo | Makefile alias |
| `make coin-executor` | `coin/coin-executor` | job `coin-executor` |
| `make agents-build` | — | job `agents-build` |
| `make samples` | `coin/demo-*` | multibranch job на каждый demo-репо |

Повторный вызов platform-команд **обновляет** код в Gitea. Перед push подтягиваются CI-артефакты из Gitea (**только если файл не меняли локально** с прошлого push; при конфликте побеждает monorepo):

| Команда | Sync из Gitea → monorepo | State file |
|---------|--------------------------|------------|
| `make coin-jenkins-agents` | `catalog.yaml` | `docker/.coin-jenkins-agents-sync.sha256` |

JCasc reload — у `coin-executor`, `agents-build`, `samples`.

### Demo-продукты

Манифест: `samples.yaml`. Локальные git-клоны: `../samples/` (gitignored в monorepo).

1. Копирует starter → `samples/<repo>/`, push в Gitea
2. Создаёт **multibranch job** в Jenkins (имя = имя репо)
3. Jenkins сканирует ветки по расписанию (15m) или вручную («Scan Repository Now»)

```bash
make samples    # starters → samples/<repo>/ + Gitea + Jenkins multibranch jobs
```

Можно работать как с обычными репо: ветки, коммиты, `git push origin …`.

---

## Gitea-репозитории

**Platform**

- `coin/coin-jenkins-agents`
- `coin/coin-starters`
- `coin/coin-executor`

**Product demos** (после `make samples`)

- `coin/demo-go-app`, `coin/demo-python-uv`, `coin/demo-python-pip`, `coin/demo-java-maven`, `coin/demo-java-gradle`

Каталог `docker/` в monorepo **не** попадает в Gitea.

---

## Архитектура

```
bootstrap → Jenkins + k3s + Nexus + Gitea
     │
     ├─ make coin-executor / agents-build  →  platform jobs
     ├─ make coin-platform                 →  GP + agents catalog
     └─ make samples                       →  product repos + multibranch jobs

service pipeline (Jenkinsfile.coin):
  resolve manifest → dynamic agent → coin-executor stages
```

Pod-сеть k3s не видит docker-compose DNS. Bootstrap + `fix-k3s.sh` регистрируют Service/Endpoints для `gitea`, `jenkins`, `nexus`.

---

## Структура `docker/`

```
compose.yml
samples.yaml                  # манифест demo-репозиториев

jenkins/                      # bootstrap: образ Jenkins
  Dockerfile, plugins.txt, casc.yaml, jenkins-entrypoint.sh

platform/jenkins/             # JCasC (make coin-executor / agents-build)
  casc-agents-build.yaml

k3s/
  registries.yaml             # insecure registry nexus:8082
  dashboard/                  # manifests для make dashboard
  jenkins-sa.yaml

scripts/
  bootstrap.sh                # make bootstrap
  prepare-gitea.sh
  prepare-nexus.sh
  setup-jenkins-k8s-auth.sh
  register-jenkins-k8s-endpoints.sh
  fix-k3s.sh
  setup-dashboard.sh
  coin-platform.sh
  coin-executor.sh
  agents-build.sh
  samples.sh
  reset.sh, teardown.sh
  detect-platform.sh
  lib/common.sh
```

---

## Переменные окружения

Файл `.env` (из `.env.example`). Основные:

| Переменная | Default | Назначение |
|------------|---------|------------|
| `DOCKER_PLATFORM` | `linux/arm64` | платформа образов (auto-detect в bootstrap) |
| `JENKINS_HTTP_PORT` | `8080` | Jenkins UI |
| `GITEA_HTTP_PORT` | `3000` | Gitea UI / git |
| `NEXUS_HTTP_PORT` | `8081` | Nexus UI |
| `NEXUS_DOCKER_PORT` | `8082` | Docker registry |
| `NEXUS_ADMIN_PASSWORD` | `coin12345` | Nexus admin |
| `NEXUS_IMAGE_TAG` | `3.91.1-alpine` | multi-arch Nexus |

После смены `.env`: `docker compose up -d --force-recreate <service>` или `make bootstrap`.

---

## Troubleshooting

**Bootstrap долго на Jenkins plugins** — 5–15 мин при первой сборке. Повторный bootstrap пропускает build если образ есть.

**OOM (exit 137)** — увеличьте RAM Docker (8 GB+). Dashboard только через `make dashboard`.

**Plugin download `Truncated chunk`** — retry в Dockerfile; повторите `make build-jenkins`.

**Nexus 403 EULA** — `./scripts/prepare-nexus.sh` или `make reset && make bootstrap`.

**Apple Silicon Nexus** — `NEXUS_IMAGE_TAG=3.91.1-alpine`, `DOCKER_PLATFORM=linux/arm64`.

**Jenkins pod / k8s auth** — `./scripts/setup-jenkins-k8s-auth.sh` + `docker compose up -d jenkins`.

**Gitea/Jenkins unreachable from pod** — `make bootstrap` или `./scripts/fix-k3s.sh`.

**Stale Jenkins volume / CASC** — `make reset && make bootstrap`.

**Gitea clone failed** — перепушьте repos:

```bash
make coin-platform coin-executor
make samples
```

**Jenkins container со старыми volume mounts** — после смены `compose.yml`:

```bash
docker compose up -d --force-recreate jenkins
```

---

## Связанные документы

- [coin-jenkins-agents/README.md](../coin-jenkins-agents/README.md)
- [coin-starters/README.md](../coin-starters/README.md)
- [docs/jenkins-setup.md](../docs/jenkins-setup.md)
- [docs/agent-build-model.md](../docs/agent-build-model.md)
- [docs/architecture.md](../docs/architecture.md)
