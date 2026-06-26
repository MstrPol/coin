# OpenAPI / Swagger (coin-api v1)

**Ticket:** P2-08

## Спецификация

Канонический контракт: [`coin-api/openapi/v1.yaml`](../coin-api/openapi/v1.yaml)

## Просмотр (Swagger UI)

**Встроенный UI** (coin-api):

- Прямой доступ: http://localhost:8090/docs/
- Через coin-ui proxy: http://localhost:8091/api/docs/
- Ссылка **API docs** в header operator console

Спецификация YAML: `GET /openapi/v1.yaml`

```bash
# альтернатива без coin-api:
npx @redocly/cli preview-docs coin-api/openapi/v1.yaml
```

## Генерация типов для coin-ui

```bash
cd coin-ui && make openapi-ui
# → src/api/schema.d.ts
```

## Endpoints (summary)

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| GET | `/ready` | — | Readiness |
| GET | `/v1/golden-paths/{name}/versions/{ver}/manifest` | Bearer | Resolve manifest |
| POST | `/v1/builds/report` | Bearer | Build telemetry |
| GET | `/v1/admin/me` | X-API-Key / Bearer | Current user roles |
| GET | `/v1/admin/projects` | X-API-Key | Project list (`?stale=true`) |
| GET | `/v1/admin/build-reports` | X-API-Key | Build reports list |
| GET | `/v1/admin/golden-paths` | X-API-Key | GP releases |
| GET | `/v1/admin/golden-paths/{name}/versions/{ver}` | X-API-Key | GP release detail + composition |
| POST | `/v1/admin/golden-paths/{name}/versions` | X-API-Key | Publish GP |
| GET | `/v1/admin/golden-paths/{name}/versions/{ver}/blast-radius` | X-API-Key | Blast radius |
| GET | `/v1/admin/components` | X-API-Key | Component registry list |
| GET | `/v1/admin/components/{type}/{name}` | X-API-Key | Component detail |
| GET | `/v1/admin/components/{type}/{name}/versions/{ver}` | X-API-Key | Version metadata |
| GET | `/v1/admin/components/agent/{name}/next-version?runtime=` | X-API-Key | Legacy stack agents `{runtime}-r{N}`; `coin-agent` — semver через publish-agent |
| POST | `/v1/admin/components/{type}/{name}/versions` | X-API-Key | Publish component |
| GET | `/v1/admin/golden-paths/{name}/resolve-preview` | X-API-Key | Resolve debug |
| GET | `/v1/admin/golden-paths/{name}/projects/{project}/canary-context` | X-API-Key | Canary audience |
| GET | `/v1/admin/audit-log` | X-API-Key | Audit log list |
| GET | `/v1/admin/golden-paths/names` | X-API-Key | GP names for wizard |
| GET | `/v1/admin/golden-paths/{name}/profile` | X-API-Key | Composition slots |
| GET | `/metrics` | — | Prometheus |

JSON Schema manifest: [`coin-api/manifest.schema.json`](../coin-api/manifest.schema.json)
