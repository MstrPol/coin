# coin-executor charter

## In scope

| Command | Role |
|---------|------|
| `validate` | Product `.coin/config.yaml` vs schema from manifest |
| `run` | Execute pipeline stages from manifest |
| `version` | Application version from git tags |
| `report` | POST build telemetry to coin-api |

## Out of scope

- GP publish / composition
- Project scanner / analytics
- Admin API
- Dockerfile templating engine (script stage in git content)

**Stateless:** one manifest + one project config = one pipeline run.
