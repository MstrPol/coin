# Starters

Скелетоны новых product repos — **отдельный git repo** (`coin-starters`, PF-16).

Это **не** golden path runtime. Starter — one-time copy / `coin init` в product repo.

```bash
cp -r coin-starters/go-app/ my-service/
cp coin-starters/Jenkinsfile.coin my-service/Jenkinsfile
```

| Starter | GP name (config) |
|---------|------------------|
| `go-app/` | `go-app` |
| `java-gradle-app/` | `java-gradle-app` |
| `java-maven-app/` | `java-maven-app` |
| `python-uv-app/` | `python-uv-app` |
| `python-pip-app/` | `python-pip-app` |

Каждый starter: `.coin/config.yaml`, thin `Jenkinsfile` (`coinPipeline()`), минимальный код.

Эталон thin Jenkins: [`Jenkinsfile.coin`](Jenkinsfile.coin) — `@Library('coin-lib@1.0.0')` + `coinPipeline()`; stages из resolved manifest.

Local pilot: `cd docker && make coin-starters` → Gitea `coin/coin-starters`.
