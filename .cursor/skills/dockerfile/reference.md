# Dockerfile — справочник (Habr 1041784)

Источник: [Best Practices по Dockerfile](https://habr.com/ru/articles/1041784/) (casssuzy, DevOps/DevSecOps).

## 1. Базовый образ

- Минимальный подходящий образ: `python:3.12-slim`, не `ubuntu` + ручная установка runtime.
- **slim** — часто лучший компромисс (glibc).
- **alpine** — только если зависимости проверены на musl.
- **distroless** / **scratch** — для статических бинарников (Go и т.д.), сложнее отладка.
- Prod: фиксированный тег или `@sha256:digest`, не `latest`.
- Обновления базы — контролируемо (LTS, скан CVE), не `apt-get upgrade` в каждой сборке.
- Только нужные пакеты; `apt-get install -y --no-install-recommends`; `rm -rf /var/lib/apt/lists/*`.

## 2. Контекст сборки

**.dockerignore** (базовый набор):

```
.git
.env
.env.*
node_modules/
__pycache__/
*.pyc
.venv
dist/
build/
.cache/
.idea/
.vscode/
*.log
.aws/
.ssh/
```

Allowlist для Go/Java при необходимости: `*` + `!go.mod` + `!cmd/` …

- Явное копирование вместо слепого `COPY . .`.
- `COPY` > `ADD`; скачивание — `RUN curl` + проверка SHA256/GPG.
- Не `curl ... | sh` без проверки.

## 3. Слои, кеш, BuildKit

Порядок: `FROM` → OS deps → lock/manifest → install deps → исходники.

Объединять `RUN` с очисткой в одной инструкции.

```dockerfile
# syntax=docker/dockerfile:1.8
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt
```

CI: `buildx --cache-from/--cache-to`, периодическая чистая пересборка.

## 4. Multi-stage

Builder — toolchain и сборка; runtime — артефакт + runtime only.

Python: wheels в builder → `pip install --no-index --find-links=/wheels` в runtime, или перенос venv.

## 5. Секреты

Запрещено: секреты в `ENV`/`ARG`/`COPY` контекста.

```dockerfile
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN="$(cat /run/secrets/npm_token)" npm ci
```

## 6. Пользователь и FS

- Non-root (`USER app` / встроенный `node`).
- OpenShift: не жёсткий UID на writable dirs; `/tmp` где уместно.
- Код: read/execute для app, не write (`chmod -R a-w /app`).
- Root-only init: `gosu`/`su-exec`, не `sudo` для приложения.

## 7. CMD / ENTRYPOINT / сигналы

- Exec-форма JSON.
- `ENTRYPOINT` — основная команда; `CMD` — аргументы по умолчанию.
- Shell-скрипт: `exec "$@"` для PID 1; при зомби — `tini`.
- Один сервис на контейнер.

## 8. Runtime

- `HEALTHCHECK` в Dockerfile (Docker/Swarm); в K8s — probes в YAML.
- `EXPOSE` — документация; не лишние порты.
- Конфиги окружений — не bake prod config в образ.
- Логи в stdout/stderr.
- Gunicorn: `--worker-tmp-dir /dev/shm`.

## 9. Supply chain и CI

- OCI labels (`org.opencontainers.image.*`).
- Private registry, сканирование (Trivy и др.).
- Подпись: Cosign / Notation; verify в deploy.
- SBOM/provenance: `docker buildx build --sbom=true --provenance=true`.
- Линтеры: `hadolint`, `docker buildx build --check .`.
- Теги: версия + `git-${SHA}`, не только `latest`.
- Не монтировать `/var/run/docker.sock` без крайней нужды.

## 10. Примеры из статьи

### Python API (сокращённо)

```dockerfile
# syntax=docker/dockerfile:1.8
FROM python:3.12.13-slim-bookworm AS builder
WORKDIR /build
RUN apt-get update && apt-get install -y --no-install-recommends build-essential \
    && rm -rf /var/lib/apt/lists/*
COPY requirements.txt ./
RUN --mount=type=cache,target=/root/.cache/pip \
    pip wheel --wheel-dir /wheels -r requirements.txt

FROM python:3.12.13-slim-bookworm AS runtime
WORKDIR /app
RUN groupadd -r app && useradd -r -g app app
COPY --from=builder /wheels /wheels
COPY requirements.txt ./
RUN pip install --no-cache-dir --no-index --find-links=/wheels -r requirements.txt && rm -rf /wheels
COPY . /app
RUN chmod -R a-w /app
USER app
EXPOSE 8000
ENTRYPOINT ["gunicorn", "--worker-tmp-dir", "/dev/shm", "app.main:app", "-k", "uvicorn.workers.UvicornWorker", "-b", "0.0.0.0:8000"]
```

### Node.js (сокращённо)

```dockerfile
# syntax=docker/dockerfile:1.8
FROM node:24.16.0-slim AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --omit=dev

FROM node:24.16.0-slim AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=deps /app/node_modules ./node_modules
COPY package.json ./
COPY src/ ./src/
RUN chmod -R a-w /app
USER node
EXPOSE 3000
CMD ["node", "src/index.js"]
```

## Чеклист перед merge

- [ ] Версия базового образа зафиксирована
- [ ] `.dockerignore` актуален
- [ ] Multi-stage для compiled/runtime separation
- [ ] Non-root, exec CMD/ENTRYPOINT
- [ ] Нет секретов в слоях
- [ ] Кеш-friendly порядок COPY/RUN
- [ ] `# syntax=docker/dockerfile:1.8` при cache/secret mounts
- [ ] Labels/version для runtime-образов Coin
- [ ] Согласованность с `images.yaml` / шаблонами config
