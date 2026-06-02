# Архитектура Coin

## Принцип разделения

```
┌─────────────────────────────────────────────────────────┐
│  Jenkins (coin-lib)  — ТОЛЬКО оркестрация               │
│  выбор агента, credentials, стадии, QG-пороги           │
├─────────────────────────────────────────────────────────┤
│  coin CLI (в образе агента)  — ВСЯ логика               │
│  валидация, версионирование, сборка, публикация          │
└─────────────────────────────────────────────────────────┘
```

Jenkins **не вычисляет** версии, **не рендерит** Dockerfile, **не знает** о деталях стеков.
Он вызывает `coin` и получает результат.

## Компоненты

| Компонент | Язык | Назначение |
|-----------|------|------------|
| `coin-lib` | Groovy | Тонкий Jenkins-оркестратор |
| `coin-cli` | Go | CLI с встроенными скриптами и шаблонами |
| `coin-images` | Docker | Toolchain-образы агентов (содержат coin CLI) |
| `coin-templates` | — | Golden path шаблоны для новых сервисов |

## coin CLI

Единый Go-бинарь без внешних зависимостей. Скрипты и Dockerfile-шаблоны
вшиты через `go:embed` — образ агента получает один файл и он самодостаточен.

### Команды

```
coin validate                                  # валидация .coin/config.yaml
coin version                                   # показать текущую версию (read-only)
coin version bump patch|minor|major            # создать следующий snapshot-тег
coin version bump patch|minor|major --type rc  # создать следующий RC-тег (только release/*)
coin version bump patch --dry-run              # показать тег без создания
coin run test|build|publish                    # запустить стандартную стадию
coin dockerfile render                         # сгенерировать managed Dockerfile
```

#### `coin version` — read-only

Выводит последнюю версию из git-тегов. Нет тегов → `0.0.1`.

```bash
$ coin version
1.5.0-PROJ-404-rc-2          # HEAD помечен RC-тегом
0.0.1-PROJ-101-snapshot-3    # последний snapshot-тег в репо
0.0.1                        # тегов нет (новый проект)
```

Используется в CI для инъекции версии в сборку:

```groovy
stage('Build') {
    sh 'COIN_VERSION=$(coin version) coin run build'
}
```

#### `coin version bump` — создание тега

Вычисляет и создаёт следующий тег. По умолчанию `--type snapshot`.

```bash
coin version bump patch            # → v0.0.1-PROJ-101-snapshot-1 (новая серия)
coin version bump patch            # → v0.0.1-PROJ-101-snapshot-2 (продолжение серии)
coin version bump minor --type rc  # → v0.1.0-PROJ-404-rc-1 (только на release/*)
coin version bump minor --type rc  # → v0.1.0-PROJ-404-rc-2 (итерация ПСИ)
```

Логика выбора базовой версии:
- Для данного (JIRA-ID, тип) уже есть серия → продолжить её (same base, N+1).
- Нет серии → взять последний base из любых тегов, применить bump, N=1.

### Как попадает в образ агента

```
coin-cli CI  →  Go build  →  coin_linux_amd64  →  Nexus/Artifactory
                                                        │
                                    coin-images/*/Dockerfile
                                    ADD <nexus>/coin-cli/${COIN_CLI_VERSION} /usr/local/bin/coin
```

Версия CLI фиксируется в `coin-lib/resources/images.yaml` рядом с версией образа.

### Локальный запуск

```bash
coin validate        # проверить config.yaml
coin version         # посмотреть что будет COIN_VERSION на текущей ветке
coin run test        # запустить тесты локально как в CI
```

## Как Jenkins вызывает CLI

`coinPipeline.groovy` (целевое состояние):

```groovy
// Стандартный пайплайн (feature/bugfix/release ветки):
stage('Validate') { sh 'coin validate' }
stage('Test')     { sh 'coin run test' }
stage('Build')    { sh 'COIN_VERSION=$(coin version) coin run build' }
stage('Publish')  {
    withCredentials([...]) { sh 'COIN_VERSION=$(coin version) coin run publish' }
}

// coinRelease job (только release/* → создать RC-тег):
stage('Tag RC') {
    sh 'coin version bump ${BUMP_LEVEL} --type rc'
    // BUMP_LEVEL = patch | minor | major — параметр Jenkins job
}
```

Groovy отвечает только за: выбор K8s pod template, binding credentials, управление стадиями, QG-пороги.

## Совместимость lib ↔ CLI

`coin-lib` объявляет минимальную совместимую версию CLI через переменную `COIN_MIN_CLI_VERSION`.
При старте пайплайна `coin validate` проверяет версию бинаря — падает с понятным сообщением
если образ устарел.

## Миграция (переходный период)

Текущий `coin-lib` и Groovy-логика продолжают работать. CLI внедряется постепенно:

1. CLI добавляется в образы агентов.
2. Groovy-команды заменяются на `sh 'coin run ...'` поэтапно.
3. Дублирующая Groovy-логика удаляется из `coin-lib`.

## Хранилище бинарей

Бинари публикуются в корпоративный Nexus/Artifactory:
```
artifacts.company.local/coin-cli/<version>/coin_linux_amd64
artifacts.company.local/coin-cli/<version>/coin_darwin_amd64   # для локального запуска
```

Сборка и публикация — через CI самого `coin`-монорепо (`Jenkinsfile` в корне).
