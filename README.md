# Coin CI

Платформа сборки на Jenkins: оркестрация в **coin-lib**, toolchain в **coin-images**, эталоны в **coin-templates**.

## Структура

| Каталог | Назначение |
|---------|------------|
| `coin-lib/` | Jenkins Shared Library — оркестрация, стандартные сценарии, managed Dockerfile |
| `coin-images/` | Docker-образы для K8s dynamic agents |
| `coin-templates/` | Golden paths для новых сервисов |

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

- [Настройка Jenkins](docs/jenkins-setup.md)
- [Схема config.yaml](docs/config.md)
- [Разделение ответственности](docs/responsibilities.md)
- [coin-lib](coin-lib/README.md)
- [Шаблон python-uv](coin-templates/python-uv/README.md)
- [Шаблон python-pip](coin-templates/python-pip/README.md)
- [Шаблон go](coin-templates/go/README.md)
- [Шаблон java-maven](coin-templates/java-maven/README.md)
- [Шаблон java-gradle](coin-templates/java-gradle/README.md)

## Содержимое репозитория

```
coin/
├── coin-lib/           # Jenkins Shared Library
├── coin-images/        # Docker-образы для поддерживаемых stack
├── coin-templates/     # Golden paths
├── docs/
└── Jenkinsfile         # сборка образов в этом репо
```
