# Local dev: Control Plane v2

**Цель:** поднять стенд, получить manifest, прогнать demo-go-app E2E.

**Gate:** P0 go/no-go.

## Prerequisites

- Docker Desktop (running)
- `make`, `curl`, `jq` (optional)

## Шаги

```bash
cd docker
cp .env.example .env    # если нет .env
make bootstrap          # nexus, k3s, gitea, jenkins, postgres, coin-api
make endpoints          # k3s Endpoints — обязательно после up/restart
```

Verify infra:

```bash
curl -sf http://localhost:8090/ready
curl -sf http://localhost:8080/login -u admin:admin -o /dev/null
```

Platform + product:

```bash
make coin-jenkins-agents
# executor binary в Nexus (локально или Jenkins job coin-executor PUBLISH=true):
cd ../coin-executor && GOOS=linux GOARCH=arm64 go build -o /tmp/coin-executor ./cmd/coin-executor
curl -u admin:coin12345 -X PUT --upload-file /tmp/coin-executor \
  "http://localhost:8081/repository/maven-releases/coin/executor/coin-executor/0.1.0/coin-executor-0.1.0-linux-arm64"

make samples            # demo-go-app → Gitea + Jenkins multibranch
```

Resolve manifest:

```bash
curl -fsS http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest | jq '.goldenPath, .runtime, .orchestration.url'
```

Nexus fallback (после resolve прогрел cache):

```bash
SNAPSHOTS=http://localhost:8081/repository/maven-snapshots
curl -fsS "${SNAPSHOTS}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.0.json" | jq .
```

E2E smoke (PF-11):

```bash
make e2e-mvp1
```

Jenkins: http://localhost:8080 → job **demo-go-app** / **main** → Build Now → SUCCESS.

## Verify (acceptance)

- [ ] `/ready` → `{"status":"ready"}`
- [ ] manifest содержит `goldenPath.name=go-app`, `runtime.image`
- [ ] demo-go-app: validate → test → build green
- [ ] Образ `localhost:8082/coin-docker/app:<build>` создан

## Troubleshooting

| Симптом | Решение |
|---------|---------|
| Agent `offline`, JNLP Connection refused | `make endpoints` |
| `lookup nexus: no such host` при docker build | `COIN_REGISTRY_PREFIX=localhost:8082/coin-docker` (уже в Jenkinsfile.coin) |
| executor 404 | Jenkins job `coin-executor` с `PUBLISH=true` или PUT в `maven-releases/coin/executor/...` |
| manifest sha256 mismatch | `make coin-jenkins-agents` + rebuild coin-api |

## Ссылки

- [docker/README.md](../../docker/README.md)
- [onboarding-15min.md](onboarding-15min.md) — быстрый старт для новых dev
- [jenkins-setup.md](../jenkins-setup.md)
