# Starters

Скелетоны новых репозиториев — часть [coin-platform](../README.md).

Это **не** golden path. Starter — то, что команда получает через `coin init` и копирует в свой репо.

```bash
export COIN_PLATFORM_DIR=/path/to/coin-platform
coin init
```

| Starter | Golden path |
|---------|-------------|
| `go-app/` | `go-app` |
| `java-gradle-app/` | `java-gradle-app` |
| `java-maven-app/` | `java-maven-app` |
| `python-uv-app/` | `python-uv-app` |
| `python-pip-app/` | `python-pip-app` |

Каждый starter: `.coin/config.yaml`, `Jenkinsfile` (`coinPipeline()`), минимальный код и тест.
