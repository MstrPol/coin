# coin-gp-content

Immutable GP content packages per golden path stack: **build engine policy**, Containerfile, validate schema.

## Layout

```
stacks/
├── go-app/           # build.engine: buildkit
│   ├── content.yaml
│   ├── dockerfiles/Containerfile
│   └── schemas/config.v2.schema.json
├── go-app-bp/        # build.engine: buildpack
└── go-app-df/        # build.engine: dockerfile
```

`content.yaml` — SoT для `build`, typed `pipeline.stages`, `validateSchema` и GP-owned deliverables. Physical publish/cache destinations задаются в GP version, не в gp-content.

**Superseded:** `scripts/*.sh` как runtime path, отдельные pipeline/validate/dockerfile component types.

## Publish

**Primary path:** GP release detail в coin-ui — pipeline-inline v3 редактируется на draft release; promote materializes manifest.

**Seed source only:** `stacks/*` копируются в `coin-api/internal/gpcontent/seed/pipelines/` при bootstrap. Shell publish **deprecated**:

```bash
./scripts/publish-content.sh go-app 1.0.2
```

Zip → Nexus `maven-releases/coin/gp-content/{name}/{ver}/` → register в coin-api.
Не использовать как SoT для новых версий — только Studio + Admin API.

Local full stack:

```bash
cd docker
make coin-gp-content
make seed-jenkins-lib
```

## CI

Jenkins job `coin-gp-content` — `docker/scripts/coin-gp-content.sh`.

## См. также

- [docs/golden-paths.md](../docs/golden-paths.md)
- [docs/agent-build-model.md](../docs/agent-build-model.md)
