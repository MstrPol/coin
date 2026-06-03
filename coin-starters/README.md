# coin-starters

**Скелетоны новых репозиториев** — минимальный рабочий проект для старта.

Это **не** golden path. Golden path (`coin-golden-paths/`) — platform-owned профиль доставки (scripts, Dockerfile, profile).  
Starter — то, что команда **копирует в свой репозиторий**: код приложения, `Jenkinsfile`, `.coin/config.yaml`.

## Как использовать

### Интерактивный визард (рекомендуется)

```bash
coin init
```

↑↓ — выбор варианта, Enter — подтверждение, Ctrl+C — отмена.

### Неинтерактивно

```bash
coin init --yes \
  --starter python-uv-app \
  --name my-service \
  --dir ./my-service
```

### Вручную (без CLI)

```bash
cp -r coin-starters/python-uv-app/* /path/to/my-new-service/
```

## Матрица

| Starter | Golden path | Содержимое |
|---------|-------------|------------|
| [go-app/](go-app/) | `go-app` | `main.go`, `go.mod`, smoke test |
| [java-gradle-app/](java-gradle-app/) | `java-gradle-app` | Gradle, `App.java` |
| [java-maven-app/](java-maven-app/) | `java-maven-app` | Maven, `App.java` |
| [python-uv-app/](python-uv-app/) | `python-uv-app` | uv, `pyproject.toml`, pytest |
| [python-pip-app/](python-pip-app/) | `python-pip-app` | pip, `requirements.txt`, pytest |

Каждый starter включает:

- `.coin/config.yaml` — эталон из `coin-golden-paths/<name>/v1/config.yaml`
- `Jenkinsfile` — `@Library('coin-lib')` + `coinPipeline()`
- минимальный код и тест, проходящий `coin run test`

## Связь с golden path

```
coin-golden-paths/python-uv-app/v1/   ← платформа (profile, scripts, runtime-only Dockerfile)
coin-starters/python-uv-app/          ← команда копирует в свой репо
```

CI: test + build (native) в agent image; app image — pack runtime-only Dockerfile.  
Подробнее — [docs/agent-build-model.md](../docs/agent-build-model.md), [docs/golden-paths.md](../docs/golden-paths.md).
