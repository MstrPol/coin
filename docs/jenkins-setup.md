# Настройка Jenkins (Control Plane v2)

## Service pipeline

Продуктовый `Jenkinsfile` — thin bootstrap + Shared Library:

```groovy
@Library('coin-lib@1.0.0') _
coinPipeline()
```

Эталон: [`samples/demo-go-app/Jenkinsfile`](../samples/demo-go-app/Jenkinsfile).

### Flow (`coinPipeline`)

1. **Resolve** (controller node) — `coin-api` `GET /v1/golden-paths/{gp}/resolve?pin=…` → stash manifest
2. **Pod** — `podTemplate` с `manifest.runtime.image` (`coin-agent`)
3. **Checkout** — product repo
4. **Materialize** — `.coin/manifest.json`, `.coin/effective-config.yaml`
5. **Bootstrap** — podman system service
6. **Stages** — динамически из `manifest.pipeline.stages` → `coin-executor run --stage …`
7. **Report** — `coin-executor report` → `POST /v1/builds/report`

**Не используется:** v1 fat pipeline, `curl orchestration` из Nexus, download executor в bootstrap, `container('stack')`.

### Параметр `publish`

| `publish` | Stage `publish` |
|-----------|-----------------|
| `false` (default) | skip (coin-lib) |
| `true` | выполняется; eligibility — `manifest.branching` + `COIN_PUBLISH_REQUEST` |

См. [how-to/branching-models.md](how-to/branching-models.md), [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md).

## Kubernetes cloud

Локальный стенд: cloud `kubernetes` из JCasC (`docker/jenkins/casc.yaml`).

**После restart Docker:**

```bash
cd docker && make endpoints
```

Обновляет Endpoints `jenkins`, `nexus`, `gitea`, `coin-api` в k3s.

## Credentials

| ID | Назначение |
|----|------------|
| `nexus-docker` | Docker registry (push/pull) |
| `nexus-admin` | Platform jobs |
| `k3s-token` | Jenkins → k8s API |
| `coin-api-token` | Bearer для Resolve + Report |
| `gitea-git` | SCM product + coin-lib repos |

### Auth policy

| Окружение | `AUTH_DISABLED` | Jenkins |
|-----------|-----------------|---------|
| Local dev | `true` | credential всё равно используется |
| Prod-like | `false` | `coin-api-token` обязателен |

## Platform CI jobs / make targets

| Команда | Назначение |
|---------|------------|
| `make coin-executor` | Gitea repo + job `coin-executor` |
| `make coin-lib` | Gitea tag `1.0.0` + Global Shared Library |
| `make publish-agent` | `coin-agent` image → Nexus + coin-api |
| `make seed-jenkins-lib` | lib + branching-model + GP (embedded pipeline) |
| `make samples` | demo repos → Gitea + multibranch |
| `make e2e-build-engines` | E2E: demo-go-app, demo-go-app-docker |
| `make e2e-mvp1` | Smoke resolve + Nexus без Jenkins |

**Superseded:** `make coin-jenkins-agents`, `make coin-gp-content`, job `agents-build`.

## Prod-like стенд (quick path)

```bash
cd docker
make bootstrap && make endpoints
make publish-agent GOARCH=arm64    # Apple Silicon
make seed-jenkins-lib              # lib + branching + GP + coin-lib-http
make samples
make e2e-build-engines
```

**Deprecated:** `make coin-lib` (Gitea SCM retriever) — только для legacy bootstrap.

Verify: Jenkins → `demo-go-app`, `demo-go-app-docker` → main → SUCCESS.

## См. также

- [agent-build-model.md](agent-build-model.md) — pod, bootstrap, engines
- [how-to/troubleshoot-ci.md](how-to/troubleshoot-ci.md)
- [runbooks/api-down-nexus-fallback.md](runbooks/api-down-nexus-fallback.md)
