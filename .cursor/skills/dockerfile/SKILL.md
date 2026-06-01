---
name: dockerfile
description: >-
  Создаёт и ревьюит Dockerfile и .dockerignore по best practices (воспроизводимость,
  размер, безопасность, кеш BuildKit). Используй при создании или изменении Dockerfile,
  Containerfile, .dockerignore, stack-образов в coin-images, managed-шаблонов в
  coin-lib/resources/dockerfiles, или когда пользователь просит контейнеризацию.
---

# Dockerfile (Coin CI)

Рабочая памятка по [Best Practices Dockerfile (Habr)](https://habr.com/ru/articles/1041784/). Подробный чеклист — в [reference.md](reference.md).

## Когда какой Dockerfile

| Контекст | Путь | Назначение |
|----------|------|------------|
| Stack-образ для Jenkins/K8s agent | `coin-images/<stack>/Dockerfile` | Toolchain (Python+uv, Go, JVM), `WORKDIR /workspace`, пользователь `ci` |
| Managed runtime сервиса | `coin-lib/resources/dockerfiles/<template>/Dockerfile` | Multi-stage, плейсхолдеры `{{...}}`, не копировать в репозитории сервисов |
| Эталон для команды | `coin-lib/resources/dockerfiles/*` + `coin-lib/resources/dockerignore/*` | Синхронизировать с `coin-templates/` и `images.yaml` |

В репозиториях сервисов с `pipeline.build.target: container` Dockerfile **не хранят** — только `dockerfileTemplate` в `.coin/config.yaml` (см. `docs/responsibilities.md`).

## Workflow перед правками

1. Определи тип образа (stack vs runtime) и стек (python-uv, go, java-maven…).
2. Прочитай соседний Dockerfile того же стека в репозитории — повтори стиль (версии, UID, apt).
3. Добавь или обнови `.dockerignore` из `coin-lib/resources/dockerignore/<stack>/`.
4. Собери Dockerfile по чеклисту ниже.
5. Проверь: `hadolint <file>` (если установлен), `docker buildx build --check .` в каталоге с Dockerfile.

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

### Multi-stage (runtime сервисов)

- Builder: компиляторы, dev-зависимости, `uv sync` / `go build` / `mvn package`.
- Runtime: только артефакт + runtime; без gcc, git, исходников (если не нужны).

### Секреты

- Не `ENV`/`ARG` для токенов; не секреты в build context.
- Build-time: `RUN --mount=type=secret,id=...`.
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
| Builder-инструменты в финальном образе | multi-stage `COPY --from=builder` |

## Шаблон: stack-образ (`coin-images`)

```dockerfile
FROM python:3.13-slim-bookworm

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

## Шаблон: managed runtime (фрагмент)

```dockerfile
# syntax=docker/dockerfile:1.8
# Coin managed Dockerfile — не копировать в репозитории сервисов.

FROM ... AS builder
COPY pyproject.toml uv.lock* ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-install-project --no-dev
COPY . .
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

FROM python:{{PYTHON_VERSION}}-slim-bookworm AS runtime
# ... non-root, COPY --from=builder, EXPOSE, CMD ...
```

## После изменений в monorepo coin

- Stack-образ: обновить `coin-images/Makefile`, пересобрать, обновить `coin-lib/resources/images.yaml`.
- Шаблон runtime: проверить `coin-templates/<stack>/` и документацию в `docs/config.md`.

## Дополнительно

- Supply chain (SBOM, Cosign, Trivy, hadolint) — см. [reference.md](reference.md) § CI/CD.
- Полные примеры Python API и Node — [reference.md](reference.md) § Примеры.
