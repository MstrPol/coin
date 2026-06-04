# Coin CI

```
 ██████╗ ██████╗ ██╗███╗   ██╗
██╔════╝██╔═══██╗██║████╗  ██║
██║     ██║   ██║██║██╔██╗ ██║
██║     ██║   ██║██║██║╚██╗██║
╚██████╗╚██████╔╝██║██║ ╚████║
 ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝
```

Платформа сборки на Jenkins: оркестрация в **coin-lib**, исполнение в **coin-cli**, описание платформы в **coin-platform**.

## Структура

| Каталог | Назначение |
|---------|------------|
| `coin-lib/` | Jenkins Shared Library — оркестратор (pod, credentials, стадии) |
| `coin-cli/` | Go CLI — validate, version, run, dockerfile render |
| `coin-platform/` | GP, starters, agent images — единое описание платформы |

Подробнее — [coin-platform/README.md](coin-platform/README.md), [docs/architecture.md](docs/architecture.md).

## Быстрый старт (приложение)

1. `export COIN_PLATFORM_DIR=./coin-platform` (или путь к клону `coin/coin-platform`)
2. `coin init` — скелетон из `coin-platform/starters/`
3. Отредактируйте `.coin/config.yaml`
4. Jenkins: `@Library('coin-lib')` + `coinPipeline()`

## Быстрый старт (platform / DevOps)

```bash
cd docker && make bootstrap
make coin-lib && make coin-platform && make coin-cli
make agents-build && make samples   # опционально
```

Подробнее — [docker/README.md](docker/README.md).

## Проверка platform

```bash
export COIN_PLATFORM_DIR=./coin-platform
coin platform validate
```

## Jenkinsfile (в сервисе)

```groovy
@Library('coin-lib@1') _

coinPipeline()
```

## Документация

Полный индекс — [docs/README.md](docs/README.md).

- [Архитектура](docs/architecture.md)
- [coin-platform](coin-platform/README.md)
- [Настройка Jenkins](docs/jenkins-setup.md)
- [Схема config.yaml](docs/config.md)
- [Golden paths](docs/golden-paths.md)
- [coin-lib](coin-lib/README.md)

## Содержимое репозитория

```
coin/
├── coin-lib/
├── coin-cli/
├── coin-platform/
│   ├── golden-paths/
│   ├── starters/
│   └── agents/
├── docs/
```
