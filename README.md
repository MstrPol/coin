# Coin CI

```
 ██████╗ ██████╗ ██╗███╗   ██╗
██╔════╝██╔═══██╗██║████╗  ██║
██║     ██║   ██║██║██╔██╗ ██║
██║     ██║   ██║██║██║╚██╗██║
╚██████╗╚██████╔╝██║██║ ╚████║
 ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝
```

Платформа сборки на Jenkins: оркестрация в **coin-lib**, toolchain в **coin-jenkins-agents**, golden paths в **coin-golden-paths**, скелетоны в **coin-starters**.

## Структура

| Каталог | Назначение |
|---------|------------|
| `coin-lib/` | Jenkins Shared Library — тонкий оркестратор (стадии, credentials, QG) |
| `coin-cli/` | Go CLI — вся логика: версионирование, сборка, публикация, релиз |
| `coin-jenkins-agents/` | Docker-образы toolchain для Jenkins dynamic agents (K8s pod `stack`-контейнер) |
| `coin-golden-paths/` | Golden paths — профили доставки (platform-owned) |
| `coin-starters/` | Скелетоны новых репозиториев (копирует команда) |

Подробнее об архитектуре — в [docs/architecture.md](docs/architecture.md).

## Быстрый старт (приложение)

1. `coin init` или скопируйте скелетон из `coin-starters/` — [coin-starters/README.md](coin-starters/README.md).
2. Отредактируйте `.coin/config.yaml` (`project.name`, credentials).
3. Зарегистрируйте Global Pipeline Library в Jenkins — [docs/jenkins-setup.md](docs/jenkins-setup.md).
4. Platform собирает agent images и coin-cli (Jenkins jobs monorepo `coin`). Обновите `coin-lib/resources/images.yaml` после релиза agents.

## Быстрый старт (platform / DevOps)

1. Platform CI: job `coin-cli` + `coin-agents` — [docs/jenkins-setup.md](docs/jenkins-setup.md).
2. Локальный стенд: `cd docker && make bootstrap` — [docker/README.md](docker/README.md).
3. Модель сборки сервисов: [docs/agent-build-model.md](docs/agent-build-model.md).

## Jenkinsfile (в сервисе)

```groovy
@Library('coin-lib@1') _

coinPipeline()
```

Конфигурация — в `.coin/config.yaml`.

## Документация

Полный индекс — [docs/README.md](docs/README.md).

- [Архитектура](docs/architecture.md)
- [Настройка Jenkins](docs/jenkins-setup.md)
- [Схема config.yaml](docs/config.md)
- [Модель ветвления](docs/branching.md)
- [Разделение ответственности](docs/responsibilities.md)
- [Golden paths](docs/golden-paths.md)
- [Модель сборки agent + runtime](docs/agent-build-model.md)
- [Скелетоны проектов](coin-starters/README.md)
- [Версионирование golden paths](docs/golden-path-versioning.md)
- [coin-lib](coin-lib/README.md)

## Поддерживаемые топологии репозиториев

Coin поддерживает **только polyrepo**: одно репо = один проект = один артефакт.

Это осознанное решение: модель ветвления, версионирование (тег = версия одного дистрибутива) и релизный процесс (ПСИ → approvals → прод) спроектированы под этот подход. Он проще в поддержке, Jenkins multibranch работает с ним нативно.

> **Монорепо (1 репо = N проектов) — не поддерживается.**
>
> Поддержка монорепо потребует отдельной архитектуры: детектирование изменений по path,
> независимые версии и теги на каждый модуль, матрица сборок.
> Работы в этом направлении начнутся только при наличии **практических кейсов**,
> где такой подход реально оправдан (атомарные изменения нескольких артефактов,
> которые нельзя решить через versioned shared library).
>
> Если у вашей команды есть такой кейс — опишите его и подайте запрос в платформенную команду.

---

## Содержимое репозитория

```
coin/
├── coin-lib/           # Jenkins Shared Library (оркестратор)
├── coin-cli/           # Go CLI (вся исполняемая логика)
├── coin-jenkins-agents/  # Docker-образы Jenkins dynamic agents (toolchain)
├── coin-golden-paths/  # Golden paths (profile, scripts, Dockerfile)
├── coin-starters/      # Скелетоны новых репозиториев
├── docs/
```
