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
make bootstrap                # infra: nexus, k3s, gitea, jenkins

make coin-lib                 # Shared Library → Gitea + Jenkins
make coin-platform            # GP + starters + agents → Gitea
make coin-cli                 # CLI → Gitea + job coin-cli
make agents-build             # job agents-build (JCasC)
make samples                  # demo-продукты → samples/ + Gitea + multibranch jobs
```

| URL | Логин |
|-----|-------|
| http://localhost:8080 | Jenkins — `admin` / см. `.env` |
| http://localhost:3000 | Gitea — `coin` / см. `.env` |
| http://localhost:8081 | Nexus — `admin` / см. `.env` |
| https://localhost:8443 | k3s Dashboard — `make dashboard` |

Docker registry: `localhost:8082/coin-docker` (`coin` / `coin1234`).

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
| `make coin-lib` | `coin/coin-lib` | Global Pipeline Library |
| `make coin-platform` | `coin/coin-platform` | — (sync CI-файлов из Gitea, если не трогали локально) |
| `make coin-cli` | `coin/coin-cli` | job `coin-cli` |
| `make agents-build` | — | job `agents-build` |
| `make samples` | `coin/demo-*` | multibranch job на каждый demo-репо |

Повторный вызов platform-команд **обновляет** код в Gitea. `make coin-platform` перед push подтягивает из Gitea CI-артефакты (`agents/catalog.yaml` и др. из списка sync), **только если файл не меняли локально** с прошлого push; при конфликте побеждает monorepo. Состояние — `docker/.coin-platform-sync.sha256` (gitignored). JCasC reload — у `coin-lib`, `coin-cli`, `agents-build`, `samples`.

### Demo-продукты

Манифест: `samples.yaml`. Локальные git-клоны: `../samples/` (gitignored в monorepo).

1. Копирует starter → `samples/<repo>/`, push в Gitea
2. Создаёт **multibranch job** в Jenkins (имя = имя репо)
3. Jenkins сканирует ветки по расписанию (15m) или вручную («Scan Repository Now»)

Требует **`make coin-lib`** (Shared Library для `Jenkinsfile`).

```bash
make samples    # starters → samples/<repo>/ + Gitea + Jenkins multibranch jobs
```

Можно работать как с обычными репо: ветки, коммиты, `git push origin …`.

---

## Gitea-репозитории

**Platform**

- `coin/coin-lib`
- `coin/coin-platform`
- `coin/coin-cli`

**Product demos** (после `make samples`)

- `coin/demo-go-app`, `coin/demo-python-uv`, `coin/demo-python-pip`, `coin/demo-java-maven`, `coin/demo-java-gradle`

Каталог `docker/` в monorepo **не** попадает в Gitea.

---

## Архитектура

```
bootstrap → Jenkins + k3s + Nexus + Gitea
     │
     ├─ make coin-lib / coin-cli / agents-build  →  platform jobs
     ├─ make coin-platform                       →  GP + agents catalog
     └─ make samples                             →  product repos + multibranch jobs

service pipeline (coinPipeline):
  checkout project → clone coin-platform → dynamic agent → coin CLI
```

Pod-сеть k3s не видит docker-compose DNS. Bootstrap + `fix-k3s.sh` регистрируют Service/Endpoints для `gitea`, `jenkins`, `nexus`.

---

## Структура `docker/`

```
compose.yml
PodTemplate.local.groovy      # override для coin-lib (docker.sock в pod)
samples.yaml                  # манифест demo-репозиториев

jenkins/                      # bootstrap: образ Jenkins
  Dockerfile, plugins.txt, casc.yaml, jenkins-entrypoint.sh

platform/jenkins/             # JCasC (make coin-lib / coin-cli / agents-build)
  casc-coin-lib.yaml
  casc-coin-cli-build.yaml
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
  coin-lib.sh
  coin-platform.sh
  coin-cli.sh
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
make coin-lib coin-platform coin-cli
make samples
```

**Jenkins container со старыми volume mounts** — после смены `compose.yml`:

```bash
docker compose up -d --force-recreate jenkins
```

---

## Связанные документы

- [coin-platform/README.md](../coin-platform/README.md)
- [docs/jenkins-setup.md](../docs/jenkins-setup.md)
- [docs/agent-build-model.md](../docs/agent-build-model.md)
- [docs/architecture.md](../docs/architecture.md)
