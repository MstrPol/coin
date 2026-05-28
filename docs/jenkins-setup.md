# Настройка Jenkins

## Global Pipeline Library

1. **Manage Jenkins** → **System** → **Global Pipeline Libraries** → **Add**.
2. Параметры:
   - **Name:** `coin-lib`
   - **Default version:** `main` (или тег `v1.0.0`)
   - **Retrieval method:** Modern SCM → Git
   - **Project repository:** URL монорепозитория `coin`
   - **Library path:** `coin-lib`
3. Сохранить.

## Kubernetes cloud

1. Установить плагины: **Kubernetes**, **Pipeline**, **Pipeline: Groovy**, **Git**, **Pipeline Utility Steps** (`readYaml`).
2. Настроить cloud (имя, например `kubernetes`) и namespace для агентов.
3. В `Jenkinsfile` сервиса при необходимости: `coinPipeline(cloud: 'kubernetes')`.

## Credentials для publish

| ID | Тип | Назначение |
|----|-----|------------|
| `coin-publish-nexus-pypi` | Username/Password | `UV_PUBLISH_USERNAME` / `UV_PUBLISH_PASSWORD` |

Дополнительно задайте переменную окружения Jenkins **`COIN_NEXUS_PYPI_URL`** — URL upload index Nexus PyPI.

## Multibranch Pipeline

1. New Item → **Multibranch Pipeline**.
2. Branch Sources → Git, URL репозитория сервиса.
3. Build Configuration → Script Path: `Jenkinsfile`.
4. Scan Triggers — по политике команды.

## Сборка образов Coin

На агенте с Docker:

```bash
cd coin-images
make build REGISTRY=registry.example.com/coin
make push REGISTRY=registry.example.com/coin
```

Обновите `coin-lib/resources/images.yaml`, если меняется registry или digest.

## Локальная отладка без K8s

```groovy
coinPipeline(kubernetes: false)
```

Требуется агент с установленными `uv` и Python 3.13.
