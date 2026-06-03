# `.coin/config.yaml`

Контракт между продуктовой командой и платформой Coin CI.

Поведение сборки задаётся **golden path** (`coin.template` + `templateVersion`), а не полями в конфиге проекта.

**Модель сборки `*-app`:** native compile в agent → runtime-only Dockerfile → registry. См. [agent-build-model.md](agent-build-model.md).

Матрица golden paths — [golden-paths.md](golden-paths.md). Версионирование каталога — [golden-path-versioning.md](golden-path-versioning.md).

Файл разделён на две явные зоны ответственности:

| Зона | Кто читает | Что содержит |
|------|-----------|--------------|
| `jenkins:` | **Jenkins (coin-lib)** | Credentials, optional override runtime/stack |
| Всё остальное | **coin CLI** | Координаты проекта, container, pipeline overrides, RN |

---

## Эталонный пример (python-uv-app)

```yaml
version: 1

coin:
  template: python-uv-app
  templateVersion: v1

# ── Jenkins (coin-lib) ───────────────────────────────────────────────────────
jenkins:
  runtime:                       # optional override версии toolchain
    python: "3.13"
  credentials:
    docker: nexus-docker
    qgm: qgm-svc-account

# ── coin CLI ─────────────────────────────────────────────────────────────────
project:
  name: my-service
  groupId: com.example.team
  repository: Nexus_PROD

container:
  port: 8080
  command: ["python", "-m", "my_service"]

pipeline:                        # optional overrides (см. ниже)
  test:
    postCommands:
      - uv run ruff check .

rn:
  serviceUrl: https://qgm.example.com
```

---

## Секция `coin`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `coin.template` | **Да** | Имя golden path: `python-uv-app`, `go-app`, … |
| `coin.templateVersion` | Нет | Версия профиля: `v1`, `v2`. Пусто → `latest` из catalog |

Stack (`python-uv`, `go`, …) **не** задаётся в проекте — coin-lib выводит его из `coin.template` через `images.yaml`.

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
    qgm: qgm-svc-account
    nexus: nexus-maven          # для *-lib шаблонов
```

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `jenkins.credentials.docker` | **Да** | Jenkins Credential ID для Docker registry |
| `jenkins.credentials.qgm` | Нет | Credential ID для QGM API |
| `jenkins.credentials.nexus` | Нет | Credential ID для Nexus (Maven/PyPI) |
| `jenkins.runtime.*` | Нет | Override версии toolchain (ключ = `images.yaml` stacks) |
| `jenkins.stack` | Нет | Override stack (если не выводится из template) |

Runtime agent image: `images.yaml` → `stacks.<stack>.<version>` (tag `{runtime}-r{N}`, optional `digest`, `rev`).

### Credentials → env

| Назначение | Env-переменные |
|------------|----------------|
| `docker` | `COIN_REGISTRY_USER`, `COIN_REGISTRY_PASSWORD` |
| `qgm` | `QGM_USER`, `QGM_PASSWORD` |
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

## Секция `container`

Параметры для **runtime-only** managed Dockerfile (`*-app`). Dockerfile **не** хранится в репозитории — рендерится в `.coin/generated/Dockerfile` (`coin dockerfile render` / `coin run build`).

Native compile выполняется в agent **до** pack; Dockerfile только копирует артефакты (`dist/`, `.venv/`, `*.jar`).

```yaml
container:
  port: 8080
  command: ["python", "-m", "my_service"]
```

---

## Секция `pipeline` — overrides

Переопределяет дефолты из `profile.yaml` шаблона. Поля `build.target`, `dockerfileTemplate`, `publish.when` **не задаются** в проекте — они platform-owned.

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

## Секция `rn` — Release Notes

```yaml
rn:
  serviceUrl: https://qgm.example.com
  codeRepository: ssh://git@bitbucket.example.com/team/my-service.git
```

Подробнее — [release-notes.md](release-notes.md).

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
| `dockerfileTemplate` | `profile.yaml` → runtime-only Dockerfile в `coin-golden-paths/<name>/vN/` |
| `publish.when` | `profile.yaml` (override через `pipeline.publish.enabled`) |
