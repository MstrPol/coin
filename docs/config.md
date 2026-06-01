# `.coin/config.yaml`

Контракт между продуктовой командой и платформой Coin CI.

Файл разделён на две явные зоны:

- **`agent:`** — читает **Jenkins (coin-lib)**: выбор образа агента и binding credentials.
- **Остальное** — читает только **coin CLI**: версионирование, сборка, публикация.

## Пример

```yaml
version: 1

coin:
  template: python-uv
  templateVersion: "1.0.0"
  versioning:
    mode: corporate
    tagPrefix: "v"

# ── Jenkins читает только эту секцию ──────────────────────────────
agent:
  stack: python-uv             # стек → выбор образа K8s-агента
  runtime:
    python: "3.13"             # версия toolchain → выбор тега образа
  publishRegistry: nexus-docker  # Jenkins Credential ID для публикации

# ── coin CLI ───────────────────────────────────────────────────────
project:
  name: my-service

container:
  port: 8080
  command: ["python", "-m", "my_service"]

pipeline:
  test:
    enabled: true
  build:
    enabled: true
    target: container          # package | container
    dockerfileTemplate: python-uv
  publish:
    enabled: true
    when: tag                  # tag | branch | always | never
```

## Секция `agent` (Jenkins)

| Поле | Описание |
|------|----------|
| `agent.stack` | Стек: `python-uv`, `python-pip`, `java-maven`, `java-gradle`, `go`, `node` |
| `agent.runtime.*` | Версия toolchain (`python: "3.13"`, `java: "17"`, `go: "1.22"`) |
| `agent.publishRegistry` | Jenkins Credential ID (`nexus-docker`, `nexus-pypi`, …) |

## Секция `coin`

| Поле | Описание |
|------|----------|
| `coin.template` | Имя golden path (например `python-uv`) |
| `coin.templateVersion` | Версия шаблона (SemVer) — для политики обязательных обновлений |
| `coin.versioning.mode` | `corporate`: версия вычисляется Coin из Git |
| `coin.versioning.tagPrefix` | Префикс релизного тега, по умолчанию `v` |

## Секция `project` и `container`

| Поле | Описание |
|------|----------|
| `project.name` | Имя сервиса (используется в image ref, Sonar) |
| `container.port` | Порт приложения — подставляется в managed Dockerfile |
| `container.command` | Команда запуска — подставляется в managed Dockerfile |

## Секция `pipeline`

| Поле | Описание |
|------|----------|
| `pipeline.<stage>.enabled` | Включить/выключить стадию |
| `pipeline.<stage>.preCommands` | Команды перед стандартным сценарием Coin |
| `pipeline.<stage>.commands` | Полная замена стандартного сценария |
| `pipeline.<stage>.postCommands` | Команды после стандартного сценария Coin |
| `pipeline.build.target` | `package` (wheel/jar/binary) или `container` (managed Dockerfile) |
| `pipeline.build.dockerfileTemplate` | Шаблон Dockerfile из Coin (например `python-uv`) |
| `pipeline.publish.when` | `tag` \| `branch` \| `always` \| `never` |

## Publish: `when`

- **`tag`** — только релизные теги `v*` (рекомендуется)
- **`branch`** — ветка `main` / `master` (snapshot-канал)
- **`always`** — каждый билд
- **`never`** — отключить

## Единое версионирование

Версию ведёт **Coin CLI**, а не сборщик проекта.
Полные правила — в [docs/branching.md](branching.md).

Coin прокидывает в сборку:
- `COIN_VERSION` — корпоративная версия артефакта
- `COIN_VERSION_SOURCE` — источник (`tag:...` или `branch:...`)
- `COIN_IMAGE_TAG` — безопасный Docker tag
- `COIN_IMAGE_REF` — полный ref образа

## Стандартные команды и расширения

По умолчанию проект **не хранит shell-скрипты**. Coin CLI запускает стандартные
сценарии, встроенные в бинарь.

Расширение тестов:
```yaml
pipeline:
  test:
    postCommands:
      - uv run ruff check .
```

Полная замена (только если стандарт не подходит):
```yaml
pipeline:
  test:
    commands:
      - uv sync --frozen --all-groups
      - uv run pytest tests/integration
```
