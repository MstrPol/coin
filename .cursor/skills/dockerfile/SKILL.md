---
name: dockerfile
description: >-
  Создаёт и ревьюит Dockerfile и .dockerignore по best practices (воспроизводимость,
  размер, безопасность, кеш BuildKit). Используй при создании или изменении Dockerfile,
  Containerfile, .dockerignore, stack-образов в coin-jenkins-agents,
  managed-шаблонов в coin-golden-paths/*-app/vN/Dockerfile,
  или когда пользователь просит контейнеризацию.
---

# Dockerfile (Coin CI)

Рабочая памятка по [Best Practices Dockerfile (Habr)](https://habr.com/ru/articles/1041784/). Подробный чеклист — в [reference.md](reference.md).

Модель сборки Coin — [docs/agent-build-model.md](../../../docs/agent-build-model.md).

## Когда какой Dockerfile

| Контекст | Путь | Назначение |
|----------|------|------------|
| CI agent (Jenkins stack container) | `coin-jenkins-agents/agents/<stack>/<runtime>.Dockerfile` | Toolchain + docker CLI; `WORKDIR /workspace` |
| App runtime (managed) | `coin-golden-paths/<gp>/vN/Dockerfile` | **Runtime-only**: `COPY` артефактов после native build в agent |
| App runtime (generated) | `.coin/generated/Dockerfile` | Render из GP при `coin run build` |

В репозиториях **сервисов** Dockerfile **не хранят** — см. `docs/responsibilities.md`.

### Запрещено в managed GP Dockerfile

- Builder stage (`AS builder`) — compile только в agent.
- Multi-stage с компилятором внутри Dockerfile app.

## Workflow перед правками

1. Определи тип: **agent image** vs **app runtime-only**.
2. Прочитай соседний Dockerfile того же стека.
3. Для app runtime: убедись, что `build.sh` создаёт артефакты до `COPY` (`dist/`, `.venv/`, `*.jar`).
4. Проверь: `hadolint`, сборку через `buildctl` или локально через `podman build`.

## Запрет Docker Daemon (BuildKit & Podman)

В проекте Coin **строго запрещено** использование классического Docker Daemon (команды `docker build`, `docker run` в CI-окружениях):
1. **Сборка** происходит исключительно через изолированные воркеры BuildKit (`buildctl`). Поэтому в начале каждого Dockerfile обязательно указывать актуальный синтаксис `# syntax=docker/dockerfile:1.8`.
2. **Запуск** контейнеров в рамках CI (тесты, временные сервисы) осуществляется **только** через `podman` или `podman-compose`.

## Обязательные правила (кратко)

### Базовый образ

- Специализированный образ под стек (`python`, `golang`, `eclipse-temurin`), не «голая» Ubuntu «на всякий случай».
- **Без `latest` в prod**: фиксируй тег (`python:3.13-slim-bookworm`, `golang:1.22-bookworm`) или digest.
- Предпочитай `slim` / `-bookworm`; Alpine — только если стек проверен на musl.
- Не делай `apt-get upgrade` / `apk upgrade` в Dockerfile; патчи — через обновление базового образа и пересборку.
- Пакеты: `--no-install-recommends`, очистка кеша в **том же** `RUN`.

### Контекст и копирование

- Всегда `.dockerignore` (секреты, `.git`, кеши, `node_modules`, `.venv`, тесты — по стеку).
- Избегай `COPY . .` до установки зависимостей; копируй явно: lock/manifest → install → исходники.
- **`COPY` вместо `ADD`**; URL и архивы — через `RUN` + checksum, не `ADD https://...`.

### Слои и BuildKit

- В начале Dockerfile для cache mount / secrets:
  ```dockerfile
  # syntax=docker/dockerfile:1.8
  ```
- Порядок: `FROM` → системные пакеты → lock-файлы → `RUN install` → код.
- Связанные команды — один `RUN`; мусор удалять в том же слое.
- Для pip/npm/uv/apt в CI — `--mount=type=cache,...` (см. managed `python-uv`).

### App runtime-only (managed GP)

- Compile **не** в Dockerfile — native в agent (`build.sh`).
- Dockerfile: runtime base + `COPY dist/` | `.venv/` | `*.jar`.
- Non-root, `EXPOSE`, `CMD {{APP_CMD}}`, `ARG COIN_VERSION`.

### Agent image (coin-jenkins-agents)

- Toolchain + coin CLI + docker CLI; `WORKDIR /workspace`, user `ci`.

### Секреты

- Не `ENV`/`ARG` для токенов; не секреты в build context.
- Build-time: `RUN --mount=type=secret,id=...`.
- Авторизация для пакетных менеджеров (npm, go, pip, maven): всегда используйте секреты BuildKit (`--mount=type=secret,id=auth_token`) для получения токенов (без сохранения в слоях) и `ARG` только для публичных URL реестров (например, `NPM_REGISTRY`, `GOPROXY`).
- Runtime: orchestrator (K8s Secrets, mounted files).

### Безопасность и процесс

- **Non-root**: в Coin — `uid/gid 1000` (`ci` или `app`), согласованно с кластером.
- Exec-форма: `CMD ["..."]`, `ENTRYPOINT ["..."]`; в shell-скриптах — `exec` для PID 1 и SIGTERM.
- Один основной процесс на контейнер.
- Код приложения: по возможности без права записи у runtime-пользователя (`chmod -R a-w` после `COPY`); writable — только `/tmp`, volume, явные каталоги.
- `EXPOSE` только нужные порты; `HEALTHCHECK` — если образ сам по себе в Docker/Swarm (в K8s — probes в манифесте).

### Метаданные (runtime)

```dockerfile
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
```

### Managed-шаблоны Coin

- Плейсхолдеры: `{{PYTHON_VERSION}}`, `{{APP_PORT}}`, `{{APP_CMD}}` — не хардкодить в шаблоне то, что приходит из config.
- Сохраняй комментарий в шапке: managed Dockerfile, не копировать в сервисы.
- Паттерн python-uv: `uv sync` с cache mount → runtime `COPY --from=builder` только `.venv` и `src`.

## Антипаттерны (отклонять в ревью)

| Плохо | Лучше |
|-------|--------|
| `FROM node:latest` | `FROM node:24.16.0-slim` |
| `COPY . .` сразу | manifest/lock → install → `COPY src/` |
| `ADD url` | `RUN curl` + `sha256sum -c` |
| `ENV API_KEY=...` | secret mount / runtime secret |
| `USER root` в runtime | `USER app` / `USER ci` |
| `CMD python app.py` (shell) | `CMD ["python", "app.py"]` |
| Builder-инструменты в **app** runtime-only Dockerfile | compile в agent; в Dockerfile только `COPY` артефактов |
| Builder-инструменты в **agent** image | toolchain в `coin-jenkins-agents/` (не в app Dockerfile) |

## Шаблон: Jenkins agent stack-образ (`coin-jenkins-agents`)

```dockerfile
# syntax=docker/dockerfile:1.8
ARG REGISTRY=""
FROM ${REGISTRY}python:3.13-slim-bookworm

ARG UV_VERSION=0.6.14
ENV UV_LINK_MODE=copy \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir "uv==${UV_VERSION}"

WORKDIR /workspace
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace
USER ci
```

## Шаблон: managed runtime-only (python-uv-app)

```dockerfile
# syntax=docker/dockerfile:1.8
# Coin managed — не копировать в репозитории сервисов.
# Артефакты (.venv, src) создаёт build.sh в agent до docker build.

ARG REGISTRY=""
FROM ${REGISTRY}python:{{PYTHON_VERSION}}-slim-bookworm
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app
COPY .venv /app/.venv
COPY src /app/src
ENV PATH="/app/.venv/bin:$PATH" COIN_VERSION="${COIN_VERSION}"
USER app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
```

## После изменений в monorepo coin

- Agent image: Jenkins job `agents-build` → версия в coin-api registry, push repo через `make coin-jenkins-agents`.
- GP runtime Dockerfile: `coin-golden-paths/<name>/vN/Dockerfile` + `scripts/build.sh`.

## Дополнительно

- Supply chain (SBOM, Cosign, Trivy, hadolint) — см. [reference.md](reference.md) § CI/CD.
- Полные примеры Python API и Node — [reference.md](reference.md) § Примеры.
