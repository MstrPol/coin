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
coin validate                               # валидация .coin/config.yaml
coin version                                # вычислить COIN_VERSION из git
coin run test|build|publish                 # запустить стандартную стадию
coin dockerfile render                      # сгенерировать managed Dockerfile
coin release bump --type patch|minor|major  # поднять тег и запушить
```

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
coin version         # посмотреть что будет COIN_VERSION
coin run test        # запустить тесты локально как в CI
```

## Как Jenkins вызывает CLI

`coinPipeline.groovy` (целевое состояние):

```groovy
stage('Validate') { sh 'coin validate' }
stage('Test')     { sh 'coin run test' }
stage('Build')    { sh 'coin run build' }
stage('Publish')  {
    withCredentials([...]) { sh 'coin run publish' }
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
