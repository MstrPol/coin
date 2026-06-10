# Настройка Jenkins (Control Plane v2)

## Service pipeline — thin bootstrap

Продуктовый `Jenkinsfile` — ~50 строк: **resolve manifest** → **load orchestration** из Nexus.

Эталон: [`samples/demo-go-app/Jenkinsfile`](../samples/demo-go-app/Jenkinsfile) (копия в [`coin-starters/Jenkinsfile.coin`](../coin-starters/Jenkinsfile.coin)).

**Не используется:** v1 Shared Library pipeline (`coinPipeline()`), git clone `coin-platform` для content.

### Flow

1. **Resolve** — `coin-api` → `.coin/manifest.json` (fallback: Nexus pointer → blob + hash verify)
2. **Orchestration** — `curl manifest.orchestration.url` → `load coinPipeline.groovy`
3. **Pod** — образ из `manifest.runtime.image`
4. **Bootstrap** — download `coin-executor` (content scripts — по URL из manifest, без git)
5. **Stages** — `coin-executor run --stage validate|test|build|publish`
6. **Report** — `coin-executor report` → `POST /v1/builds/report`

## Kubernetes cloud

Локальный стенд: bootstrap создаёт cloud `kubernetes` (см. `docker/jenkins/casc.yaml`).

**После restart Docker:**

```bash
cd docker && make endpoints
```

Обновляет Endpoints `jenkins`, `nexus`, `gitea`, `coin-api` в k3s — без этого JNLP agent offline.

## Credentials

| ID | Назначение |
|----|------------|
| `nexus-docker` | Docker push/pull |
| `nexus-admin` | Platform jobs (publish executor) |
| `k3s-token` | Jenkins → k8s API |
| `coin-api-token` | Bearer token для Resolve + Report (`COIN_API_TOKEN`) |

`gitea-git` нужен только для SCM checkout product repo (не для coin-platform content).

### Auth policy

| Окружение | `AUTH_DISABLED` | Jenkins |
|-----------|-----------------|---------|
| Local dev | `true` (default в `.env`) | credential всё равно используется |
| Prod-like / corp | `false` | `coin-api-token` **обязателен** |

CASC: [`docker/jenkins/casc.yaml`](../docker/jenkins/casc.yaml) — secret из env `COIN_API_TOKEN`.

Resolve и Report используют:

```bash
curl -H "Authorization: Bearer ${COIN_API_TOKEN}" …
```

При недоступном API — Nexus fallback только на Resolve (см. [runbooks/api-down-nexus-fallback.md](runbooks/api-down-nexus-fallback.md)).

## Platform CI jobs

| Команда | Jenkins job |
|---------|-------------|
| `make coin-executor` | `coin-executor` |
| `make agents-build` | `agents-build` |
| `make samples` | multibranch `demo-*` |
| `make e2e-mvp1` | smoke: resolve + Nexus pointer/blob/content |

## Prod-like стенд

```bash
cd docker
make bootstrap
make endpoints
make coin-jenkins-agents && make coin-starters
make coin-executor    # собрать binary → Nexus (PUBLISH=true в job)
make samples          # demo-go-app → Gitea
make e2e-mvp1         # PF-11 smoke без Jenkins
```

Verify: Jenkins job `demo-go-app` / `main` → SUCCESS (validate → test → build → report).
