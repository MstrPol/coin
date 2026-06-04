# `.coin/config.yaml`

Контракт между продуктовой командой и платформой Coin CI.

Поведение сборки задаётся **golden path** (`coin.template` + `templateVersion`), а не полями в конфиге проекта.

**Модель сборки `*-app`:** native compile в agent → runtime-only Dockerfile → registry. См. [agent-build-model.md](agent-build-model.md).

Матрица golden paths — [golden-paths.md](golden-paths.md). Версионирование каталога — [golden-path-versioning.md](golden-path-versioning.md).

Файл разделён на две явные зоны ответственности:

| Зона | Кто читает | Что содержит |
|------|-----------|--------------|
| `jenkins:` | **Jenkins (coin-lib)** | Credentials, optional override runtime/stack |
| Всё остальное | **coin CLI** | Привязка к GP, координаты проекта, optional pipeline overrides |

Версия схемы конфига **не** дублируется отдельным полем — используется `coin.templateVersion` (версия golden path).

---

## Эталонный пример (python-uv-app)

```yaml
coin:
  template: python-uv-app
  templateVersion: v1

# ── Jenkins (coin-lib) ───────────────────────────────────────────────────────
jenkins:
  runtime:                       # optional override версии toolchain
    python: "3.13"
  credentials:
    docker: nexus-docker

# ── coin CLI ─────────────────────────────────────────────────────────────────
project:
  name: my-service
  groupId: com.example.team
  repository: Nexus_PROD

pipeline:                        # optional overrides (см. ниже)
  test:
    postCommands:
      - uv run ruff check .
```

---

## Секция `coin`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `coin.template` | **Да** | Имя golden path: `python-uv-app`, `go-app`, … |
| `coin.templateVersion` | Нет | Версия профиля: `v1`, `v2`. Пусто → `latest` из catalog |

Stack (`python-uv`, `go`, …) **не** задаётся в проекте — coin-lib выводит его из GP profile (`COIN_PLATFORM_DIR/golden-paths/...`).

---

## Секция `jenkins` — Jenkins

Coin-lib читает эту секцию для выбора агента и credentials.

```yaml
jenkins:
  stack: python-uv              # optional override (обычно не нужен)
  runtime:
    python: "3.13"              # optional override
  credentials:
    docker: nexus-docker
    nexus: nexus-maven          # для *-lib шаблонов (когда появятся)
```

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `jenkins.credentials.docker` | **Да** | Jenkins Credential ID для Docker registry |
| `jenkins.credentials.nexus` | Нет | Credential ID для Nexus (Maven/PyPI) |
| `jenkins.runtime.*` | Нет | Override версии toolchain (ключ см. GP profile) |
| `jenkins.agent.image` | Нет | Явный pin образа агента (минуя catalog) |
| `jenkins.stack` | Нет | Deprecated — stack из GP profile |

Runtime agent image: `COIN_PLATFORM_DIR/agents/catalog.yaml` → `stacks.<stack>.<runtime>`.

### Credentials → env

| Назначение | Env-переменные |
|------------|----------------|
| `docker` | `COIN_REGISTRY_USER`, `COIN_REGISTRY_PASSWORD` |
| `nexus` | `NEXUS_USER`, `NEXUS_PASSWORD` |

---

## Секция `project`

```yaml
project:
  name: my-service
  groupId: com.example.team
  repository: Nexus_PROD
```

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `project.name` | **Да** | Имя сервиса / artifactId |
| `project.groupId` | **Да** | Домен команды в реестре |
| `project.repository` | **Да** | Логическое имя репозитория Nexus (RN, QGM) |

---

## Секция `pipeline` — overrides

Переопределяет дефолты из `profile.yaml` шаблона. Поля `build.type`, `container.*`, `publish.when` **не задаются** в проекте — они platform-owned.

```yaml
pipeline:
  test:
    enabled: true
    postCommands:
      - uv run ruff check .
  build:
    enabled: true
  publish:
    enabled: false
```

| Поле | Описание |
|------|----------|
| `enabled` | Включить/выключить стадию |
| `preCommands` | Команды перед стандартным сценарием Coin |
| `commands` | Полная замена стандартного сценария |
| `postCommands` | Команды после стандартного сценария |

---

## Container (managed Dockerfile)

Параметры runtime-образа (`port`, `command`) задаются в **`profile.yaml` golden path**, не в конфиге проекта.

Dockerfile **не** хранится в репозитории сервиса — рендерится в `.coin/generated/Dockerfile` (`coin dockerfile render` / `coin run build`).

Пример в `coin-golden-paths/go-app/v1/profile.yaml`:

```yaml
container:
  port: 8080
  command: ["/app/app"]
```

Native compile выполняется в agent **до** pack; Dockerfile только копирует артефакты (`dist/`, `.venv/`, `*.jar`).

---

## Release Notes (QGM)

Интеграция с QGM в pipeline **пока не включена**. Координаты артефакта уже есть в `project:`.

Когда появится QGM, URL сервиса и credentials будут на уровне **platform** (не в каждом репозитории). Подробнее — [release-notes.md](release-notes.md).

---

## Версионирование артефакта

Coin CLI передаёт в сборку:

| Переменная | Описание |
|------------|----------|
| `COIN_VERSION` | Версия из git-тега |
| `COIN_IMAGE_TAG` | Docker-тег |
| `COIN_IMAGE_REF` | Полный ref образа |
| `COIN_TEMPLATE` / `COIN_TEMPLATE_VERSION` | Golden path |

Правила — [branching.md](branching.md).

---

## Что **не** задаётся в проекте

| Поле | Где живёт |
|------|-----------|
| `build.type` | `profile.yaml` шаблона |
| `agent.stack` | `profile.yaml` + `catalog.yaml` |
| `container.port` / `container.command` | `profile.yaml` шаблона |
| `dockerfileTemplate` | `profile.yaml` → runtime-only Dockerfile в `coin-golden-paths/<name>/vN/` |
| `publish.when` | `profile.yaml` (override через `pipeline.publish.enabled`) |
| `rn.serviceUrl` | platform (QGM, когда будет включено) |
