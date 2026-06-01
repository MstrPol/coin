# Coin CI

Платформа сборки на Jenkins: оркестрация в **coin-lib**, toolchain в **coin-images**, эталоны в **coin-templates**.

## Структура

| Каталог | Назначение |
|---------|------------|
| `coin-lib/` | Jenkins Shared Library — тонкий оркестратор (стадии, credentials, QG) |
| `coin-cli/` | Go CLI — вся логика: версионирование, сборка, публикация, релиз |
| `coin-images/` | Docker-образы для K8s dynamic agents (содержат coin CLI) |
| `coin-templates/` | Golden path шаблоны для новых сервисов |

Подробнее об архитектуре — в [docs/architecture.md](docs/architecture.md).

## Быстрый старт (приложение)

1. Скопируйте `coin-templates/python-uv/` в репозиторий сервиса.
2. Зарегистрируйте Global Pipeline Library в Jenkins (см. [docs/jenkins-setup.md](docs/jenkins-setup.md)).
3. Соберите и опубликуйте образы: `make -C coin-images build push` (см. `coin-images/Makefile`).
4. Обновите `coin-lib/resources/images.yaml` при смене тегов/digest.

## Jenkinsfile (в сервисе)

```groovy
@Library('coin-lib@1') _

coinPipeline()
```

Конфигурация — в `.coin/config.yaml`.

## Документация

- [Архитектура](docs/architecture.md)
- [Настройка Jenkins](docs/jenkins-setup.md)
- [Схема config.yaml](docs/config.md)
- [Модель ветвления](docs/branching.md)
- [Разделение ответственности](docs/responsibilities.md)
- [coin-lib](coin-lib/README.md)
- [Шаблон python-uv](coin-templates/python-uv/README.md)
- [Шаблон python-pip](coin-templates/python-pip/README.md)
- [Шаблон go](coin-templates/go/README.md)
- [Шаблон java-maven](coin-templates/java-maven/README.md)
- [Шаблон java-gradle](coin-templates/java-gradle/README.md)

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
├── coin-images/        # Docker-образы для поддерживаемых stack
├── coin-templates/     # Golden paths
├── docs/
└── Jenkinsfile         # сборка CLI и образов в этом репо
```
