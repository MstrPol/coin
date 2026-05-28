# coin-lib

Jenkins Shared Library для Coin CI.

## Entry point

```groovy
@Library('coin-lib@1') _

coinPipeline()
```

## Параметры `coinPipeline`

| Параметр | По умолчанию | Описание |
|----------|--------------|----------|
| `configPath` | `.coin/config.yaml` | Путь к конфигу |
| `kubernetes` | `true` | K8s pod template |
| `cloud` | — | Имя K8s cloud в Jenkins |
| `prepareAgent` | любой агент | Label для stage Prepare (чтение config) |

## Ресурсы

- `resources/images.yaml` — stack → Docker image (CI agent)
- `resources/config.schema.json` — схема конфига
- `resources/scripts/<stack>/` — стандартные сценарии test/build/publish
- `resources/dockerfiles/<template>/` — managed Dockerfile templates
- `resources/dockerignore/<template>/` — managed `.dockerignore` для build context

Проект может расширить стандартные сценарии через `pipeline.<stage>.preCommands/postCommands`
или полностью заменить stage через `pipeline.<stage>.commands`.

## Версионирование

Версия вычисляется в Coin (`Versioning`) и прокидывается в stage как:

- `COIN_VERSION`
- `COIN_VERSION_SOURCE`
- `COIN_IMAGE_TAG`
- `COIN_IMAGE_REF`

Сборщики проектов (Gradle/Maven/uv/Go tooling) не управляют корпоративной версией.

## Классы (`src/org/coin/ci/`)

- `Config` — загрузка и валидация YAML
- `Versioning` — единая корпоративная версия из Git/Jenkins
- `StackImages` — выбор образа
- `PodTemplate` — YAML pod для K8s
- `StackExecutor` — stages Test / Build / Publish
