# coin-jenkins-agents

CI agent images для Jenkins dynamic agents.

## Структура

```
agents/{stack}/{runtime}.Dockerfile
Jenkinsfile
```

Пример multi-runtime Go:

```
agents/go/1.22.Dockerfile
agents/go/1.24.Dockerfile
```

## Сборка (job `agents-build`)

Единственный параметр — **путь к Dockerfile**:

| DOCKERFILE | component | runtime |
|------------|-----------|---------|
| `agents/go/1.22.Dockerfile` | `agent/go` | 1.22 |
| `agents/go/1.24.Dockerfile` | `agent/go` | 1.24 |
| `agents/python-uv/3.13.Dockerfile` | `agent/python-uv` | 3.13 |

Flow:

1. `GET .../agent/{stack}/next-version?runtime=` → `1.22-r{N+1}`
2. `docker build -f $DOCKERFILE` → push `ci-{stack}:{version}`
3. `POST .../agent/{stack}/versions`

Image repo: `ci-{stack}` (`java-maven` → `ci-jvm-maven`, `java-gradle` → `ci-jvm-gradle`).

Registry: `COIN_REGISTRY_PREFIX` (`localhost:8082/...` для push через host docker.sock), `COIN_REGISTRY_RUNTIME_PREFIX` (`nexus:8082/...` в metadata для k3s). Cred: `nexus-docker`.

```bash
cd docker && make agents-build
cd docker && make coin-jenkins-agents
```
