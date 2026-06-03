# Coin — локальный prod-like стенд (Docker Compose)

Поднимает на **общедоступных образах** минимальную копию prod-контура:

| Сервис | Образ | Роль |
|--------|-------|------|
| **Gitea** | `gitea/gitea` | Git (monorepo coin + demo-сервисы) |
| **Registry** | `registry:2` | Docker registry (publish образов) |
| **Nexus** | `sonatype/nexus3` | Nexus-like artifact storage |
| **k3s** | `rancher/k3s` | Kubernetes для dynamic agents |
| **Jenkins** | `jenkins/jenkins:lts-jdk17` + plugins | Platform CI + service pipelines |
| **Agents** | `coin-jenkins-agents/*` | Toolchain-образы (**собирает Jenkins**, не host) |

Prod-like поток: **Gitea → Jenkins → coin-cli в Nexus → agent images в registry → service pipeline в k3s pod**.

---

## Git в Gitea

| Репозиторий | Содержимое |
|-------------|------------|
| **`coin/coin`** | monorepo платформы (coin-cli, coin-lib, coin-jenkins-agents, …) |
| **`coin/demo-python-uv`** | demo-сервис (из `coin-starters/`) |

`make repos` пушит оба репозитория. Каталог `docker/` в monorepo **не** попадает в Gitea — это инфраструктура стенда.

---

## Требования

- Docker Desktop / Docker Engine 24+
- Docker Compose v2
- **Git** на host (push репозиториев в Gitea)
- 8+ GB RAM (k3s + Nexus + Jenkins + Gitea)
- Свободные порты: `3000` (Gitea), `8080` (Jenkins), **`5050`** (registry), `8081` (Nexus), `6443` (k3s API), **`8443`** (k3s Dashboard)

> **macOS:** порт `5000` часто занят AirPlay Receiver — по умолчанию registry на **`5050`**.  
> **Apple Silicon:** в `.env` задайте `DOCKER_PLATFORM=linux/arm64` (Intel: `linux/amd64`).

На macOS/Linux host должен быть доступен `/var/run/docker.sock` (для `coin run build` через DinD-сокет в pod).

---

## Быстрый старт

```bash
cd docker
make bootstrap
```

Скрипт:

1. Собирает образ Jenkins (Go + Docker CLI для platform jobs)
2. Поднимает registry, Nexus, k3s, Gitea
3. Создаёт Nexus raw repo `coin-cli`, Endpoints в k3s
4. Пушит **coin** monorepo и **demo-python-uv** в Gitea
5. Стартует Jenkins и запускает **platform build**:
   - `coin-cli` → Nexus `coin-cli/dev/coin_linux_<arch>`
   - `coin-agents` (BUILD_ALL) → registry `registry:5000/coin/ci-*`
6. После platform build — job **`coin-demo-python-uv`** (smoke сервиса)

Откройте **http://localhost:8080** → дождитесь зелёных **coin-cli** и **coin-agents** → **coin-demo-python-uv** → Build Now.

| Сервис | URL | Логин |
|--------|-----|-------|
| Jenkins | http://localhost:8080 | `admin` / `admin` |
| Gitea | http://localhost:3000 | `coin` / `coin` |
| Nexus | http://localhost:8081 | `admin` / `coin12345` |
| **k3s Dashboard** | https://localhost:8443 | Token (см. ниже) |

Репозитории в Gitea:
- http://localhost:3000/coin/coin
- http://localhost:3000/coin/demo-python-uv

---

## Команды

```bash
make bootstrap       # полный подъём + platform build
make down            # остановить и удалить volumes
make repos           # перепушить coin + demo в Gitea
make platform-build  # coin-cli → Nexus, agents → registry (Jenkins jobs)
make gitea           # init Gitea + k3s Endpoints
make dashboard       # Kubernetes Dashboard в k3s
make dashboard-token # токен Dashboard
make k3s-fix         # после recreate k3s
make logs            # логи compose
```

---

## Jenkins jobs

### Platform (`coin/coin` monorepo)

| Job | Jenkinsfile | Результат |
|-----|-------------|-----------|
| **coin-cli** | `coin-cli/Jenkinsfile` | `go test/build` → Nexus raw |
| **coin-agents** | `coin-jenkins-agents/Jenkinsfile` | `ci-*:{runtime}-r{N}` → registry, bump `catalog.yaml` |

Связь GP → agent — `profile.yaml` (`agent.stack`, `agent.runtime`); runtime lookup — `images.yaml` (promote: `scripts/sync-agent-images.sh`).

### Service

| Job | Репо | Назначение |
|-----|------|------------|
| **coin-demo-python-uv** | `demo-python-uv` | smoke: validate / test / build / publish |

---

## Что проверяет demo-pipeline

```
checkout  →  git clone из Gitea (K8s pod)
Validate  →  coin validate
Test      →  coin run test        (native в agent)
Build     →  coin run build       (native compile → pack → OCI image)
Publish   →  coin run publish     (push → localhost:5050)
```

Pack на local стенде: docker CLI + host `docker.sock` в pod (`PodTemplate.local.groovy`).

---

## Архитектура

```
┌──────────────┐  platform jobs   ┌──────────┐     ┌───────────┐
│ Gitea/coin   │ ───────────────► │ Jenkins  │ ──► │ Nexus     │ coin-cli/dev/
│ (monorepo)   │                  │ master   │     └───────────┘
└──────────────┘                  │ +docker  │ ──► ┌───────────┐
                                  └────┬─────┘     │ registry  │ ci-* images
                                       │           └───────────┘
                         service job   │
┌──────────────┐  checkout           ▼           ┌───────────┐
│ demo-python  │ ◄───────────  k3s pod (agent) ─►│ publish   │
└──────────────┘                                └───────────┘
```

### Git в k3s pod

Pod-сеть k3s не видит docker-compose DNS (`gitea`, `jenkins`). Bootstrap регистрирует **Service+Endpoints** в k3s с IP контейнеров — checkout из pod и JNLP-подключение агентов работают как в prod.

После recreate Jenkins или k3s:

```bash
make k3s-fix          # ghost-ноды, CoreDNS, token, Endpoints jenkins+gitea
make k8s-endpoints    # только jenkins:8080 / :50000
make gitea            # полный init Gitea + Endpoints
```

### k3s Dashboard (браузер)

Bootstrap разворачивает [Kubernetes Dashboard v2.7](https://github.com/kubernetes-retired/dashboard) в k3s (NodePort `30443` → host `:8443`).

1. Откройте **https://localhost:8443** (самоподписанный сертификат — подтвердите исключение).
2. Выберите **Token**.
3. Вставьте токен:

```bash
make -C docker dashboard-token
```

Учётка `admin-user` имеет `cluster-admin` — **только для локального стенда**.

Переустановка UI без полного bootstrap:

```bash
make -C docker dashboard
```

CLI-альтернатива: `docker compose exec k3s kubectl get pods -A`.

**401 Invalid credentials** — токен устарел (после `docker compose up -d k3s` / recreate) или скопирован не полностью. Получите новый:

```bash
make dashboard-token   # из каталога docker/
```

На macOS токен копируется в буфер автоматически. Не используйте токен из старого вывода `bootstrap` — он привязан к конкретному кластеру.

### Отличия от prod

| Prod | Local compose |
|------|---------------|
| Корп. Git (GitLab/Bitbucket) | Gitea |
| Managed K8s | k3s в контейнере |
| coin-lib из Git tag | coin-lib из Gitea `coin/coin` @ `main` (libraryPath) |
| Platform CI на corp Jenkins | coin-cli + agent images через Jenkins jobs |
| Kaniko (prod) | docker.sock + docker CLI в agent (local dev) |

---

## Структура каталога

```
docker/
  compose.yml
  gitea/app.ini
  images-local.yaml       # → coin-lib/resources/images.yaml при push-coin
  jenkins/
  k3s/registries.yaml
  k3s/dashboard/          # Kubernetes Dashboard manifests
  scripts/
    prepare-gitea.sh      # init Gitea + k3s Endpoints
    prepare-nexus.sh      # Nexus raw repo coin-cli
    push-coin.sh          # monorepo coin → Gitea
    prepare-demo.sh       # demo-python-uv → Gitea
    trigger-platform-build.sh
    bootstrap.sh
```

---

## Переменные окружения

Скопируйте `.env.example` → `.env`:

```bash
GITEA_HTTP_PORT=3000
GITEA_USER=coin
GITEA_PASSWORD=coin
JENKINS_HTTP_PORT=8080
REGISTRY_PORT=5050
DOCKER_PLATFORM=linux/arm64
```

---

## Troubleshooting

**Port 5000 already in use (macOS)** — AirPlay Receiver занимает 5000. В `docker/.env`:
```bash
REGISTRY_PORT=5050
```

**Platform amd64/arm64 mismatch (Nexus)** — старые теги (`3.76.x`) только amd64. В `docker/.env`:
```bash
NEXUS_IMAGE_TAG=3.91.1-alpine
DOCKER_PLATFORM=linux/arm64
```
Удалите закешированный amd64-образ: `docker rmi sonatype/nexus3:3.76.0`

**Platform amd64/arm64 mismatch (другие сервисы)**
```bash
DOCKER_PLATFORM=linux/arm64   # Apple Silicon
DOCKER_PLATFORM=linux/amd64   # Intel / CI
```

**Jenkins: Failed to launch pod / `Invalid DER: object is not integer`** — k3s выдаёт EC client cert, fabric8 не парсит его через CertificateCredentials. Обновите bearer token и пересоздайте Jenkins:

```bash
make k8s-auth
docker compose build jenkins && docker compose up -d jenkins
```

**Jenkins: Still waiting to schedule task / `'Jenkins' is reserved for jobs with matching label expression`** — master в режиме `EXCLUSIVE` не принимает job без label. В локальном `casc.yaml` должно быть `mode: NORMAL` и `numExecutors: 1`. Пересоздайте контейнер (не только restart):

```bash
docker compose build jenkins && docker compose up -d jenkins
```

**Jenkins: Still waiting to schedule task / Waiting for next available executor** — pipeline ещё не дошёл до k8s. Pod'ы агентов — namespace **`default`**: `docker compose exec k3s kubectl get pods -n default`. После recreate k3s: `make k8s-auth && docker compose up -d jenkins`.

**Jenkins: `isn't a valid path` / CASC with `:`** — JCasC читает **каталог**, не список через `:`. Пересоберите образ:
```bash
docker compose build --no-cache jenkins
docker volume rm coin_jenkins-home
docker compose up -d jenkins
```

**Jenkins не стартует (Stopping Jenkins / still starting up)** — k3s требует kubeconfig с client cert. Пересоберите образ:
```bash
docker compose build --no-cache jenkins
docker compose up -d jenkins
docker compose logs jenkins 2>&1 | rg -i 'severe|error|casc'
# если volume в битом состоянии:
docker compose stop jenkins
docker volume rm coin_jenkins-home
docker compose up -d jenkins
```

**Jenkins: could not clone from Gitea** — перепушьте репозитории:
```bash
make repos
```

**Pod: Failed to connect to gitea** — перерегистрируйте Endpoints:
```bash
make gitea
```

**Gitea: SSH port 22 already in use / контейнер перезапускается** — SSH отключён (git только по HTTP). Пересоздайте контейнер:
```bash
docker compose up -d --force-recreate gitea
# если не помогло — сброс volume:
docker compose down && docker volume rm coin_gitea-data && make bootstrap
```

**Gitea: read-only app.ini / migrate failed** — не монтируйте `app.ini` в `/etc/gitea`. Конфиг генерируется в volume. Сброс:
```bash
docker compose down
docker volume rm coin_gitea-data 2>/dev/null || true
make bootstrap
```

**Gitea admin / migrate**:
```bash
docker compose exec -u git gitea gitea migrate --config /data/gitea/conf/app.ini
```

---

## Связанные документы

- [docs/agent-build-model.md](../docs/agent-build-model.md)
- [docs/jenkins-setup.md](../docs/jenkins-setup.md)
- [docs/architecture.md](../docs/architecture.md)
- [docs/golden-paths.md](../docs/golden-paths.md)
