# Golden paths (Control Plane v2)

Golden path — **semver release** платформы (`go-app@1.0.0`), собранный из **backend (PostgreSQL)** + **Nexus** (published artifacts).

---

## SoT runtime (platform-first)

| Слой | Где | Что |
|------|-----|-----|
| **Authoring** | PostgreSQL `gp_artifact_bodies`, coin-ui (MVP-2) | draft/snapshot bytes |
| **Published** | Nexus `content/{gp}/{ver}/…` | scripts, Dockerfile, schema, orchestration |
| **Resolve** | coin-api → manifest JSON | url + sha256 refs, orchestration.url |

Legacy git `coin-platform/content/` **удалён** (PF-00 Gate B). Seed bytes: `coin-api/internal/gpcontent/seed/`.

---

## Composition (PostgreSQL)

GP release = набор component versions:

| Component | Пример | Роль |
|-----------|--------|------|
| `executor` | coin-executor@0.1.0 | URL binary |
| `agent` | go@1.22.5 | CI image |
| `pipeline` | go-build@2.1.0 | stage scripts bundle |
| `validate` | config@1.0.0 | JSON Schema |
| `dockerfile` | go-runtime@1.0.0 | Dockerfile template |
| `orchestration` | coin-pipeline@1.0.0 | Jenkins Groovy (`coinPipeline.groovy`) |

Seed pilot: `go-app@1.0.0` — см. `coin-api/migrations/002_seed_go_app.sql`, `009_orchestration_component.sql`.

---

## Именование

```
{stack}-{role}     →  goldenPath name
1.0.0              →  semver release (coin.version в проекте)
```

Pilot: **`go-app@1.0.0`** — seed в migrations; профили GP хранятся в PostgreSQL (`gp_profiles`, migration `010_gp_profiles.sql`).

Новый GP: `POST /v1/admin/golden-paths/profiles` (slots JSON) + publish releases как для `go-app`.

Legacy каталоги GP v1 (`profile.yaml` + git scripts) — **удалены**; v2 SoT — coin-api + Nexus.

---

## Матрица (roadmap)

| GP | v2 status | Примечание |
|----|-----------|------------|
| `go-app` | ✅ 1.0.0 | E2E demo-go-app |
| `java-*-app` | planned | P1+ |
| `python-*-app` | planned | P1+ |

---

## Resolve

```bash
curl -fsS http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest | jq .
```

Nexus (после первого resolve):

- Pointer exact pin: `http://localhost:8081/repository/coin-manifests/pointers/go-app/%3D1.0.0.json`
- Manifest blob: поле `blobUrl` в pointer
- Content: `http://localhost:8081/repository/coin-manifests/content/go-app/1.0.0/scripts/test.sh`

---

## Связанные документы

- [control-plane.md](control-plane.md)
- [config.md](config.md)
- [runbooks/api-down-nexus-fallback.md](runbooks/api-down-nexus-fallback.md)
