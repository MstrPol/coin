# Local dev: Control Plane v2

**Цель:** поднять стенд, получить manifest, прогнать E2E build engines.

**Gate:** P0 go/no-go.

## Prerequisites

- Docker Desktop (running, 20+ GB disk)
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
# Apple Silicon:
make publish-agent GOARCH=arm64
make coin-lib
make seed-jenkins-lib
make samples
```

Resolve manifest:

```bash
curl -fsS "http://localhost:8090/v1/golden-paths/go-app/resolve?pin=*" \
  | jq '{gp: .goldenPath, engine: .build.engine, runtime: .runtime.image}'
```

Nexus fallback (после resolve прогрел cache):

```bash
SNAPSHOTS=http://localhost:8081/repository/maven-snapshots
curl -fsS "${SNAPSHOTS}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.0.json" | jq .
```

E2E:

```bash
make e2e-mvp1              # smoke без Jenkins
make e2e-build-engines     # 3 jobs: buildkit, buildpack, dockerfile
```

Jenkins: http://localhost:8080 → **demo-go-app** / **main** → Build Now → SUCCESS.

## Verify (acceptance)

- [ ] `/ready` → `{"status":"ready"}`
- [ ] manifest: `goldenPath.name=go-app`, `build.engine`, `runtime.image` (coin-agent)
- [ ] demo-go-app: validate → test → build green
- [ ] Образ `localhost:8082/coin-docker/app:<build>` создан

## Troubleshooting

| Симптом | Решение |
|---------|---------|
| Agent `offline`, JNLP Connection refused | `make endpoints` |
| `lookup nexus: no such host` | Проверьте `manifest.destinations.imageRegistryPrefix` у GP release |
| manifest sha256 mismatch | Новый gp-content semver + `make seed-jenkins-lib` (Nexus immutable) |
| Pod killed (ephemeral-storage) | `bash scripts/prune-k3s-disk.sh --all` |
| buildpack pod pending | Диск + `procMount: Unmasked` в pod template |

## Ссылки

- [docker/README.md](../../docker/README.md)
- [onboarding-15min.md](onboarding-15min.md)
- [agent-build-model.md](../agent-build-model.md)
- [jenkins-setup.md](../jenkins-setup.md)
