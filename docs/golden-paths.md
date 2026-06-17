# Golden Paths (Control Plane v2)

## Модель

| Сущность | Описание |
|----------|----------|
| **Golden Path (GP)** | Именованный профиль: `go-app`, `go-app-bp`, `go-app-df`, … |
| **GP release** | Semver pin в продукте: `go-app@1.0.0` |
| **GP content** | `coin-gp-content/stacks/<gp>/` — build policy, Containerfile, schema |
| **Manifest** | JSON от Resolve: `build`, `runtime`, `pipeline`, `validateSchema` |

Продукт указывает только:

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
```

## Composition (4 slots)

При publish GP release в manifest попадают:

| Slot | Component type | Пример | Manifest |
|------|----------------|--------|----------|
| `agent` | `agent` | `coin-agent@1.0.0` | `runtime.image` |
| `executor` | `executor` | `coin-executor@0.1.0` | `executor` (binary baked в agent на pilot) |
| `lib` | `lib` | `coin-lib@1.0.0` | Jenkins `@Library` |
| `gp-content` | `gp-content` | `gp-content/go-app@1.0.2` | `build`, `pipeline`, `validateSchema` |

**Superseded:** 5-slot с `jnlp` + `agent/{stack}`, slots `pipeline` / `validate` / `dockerfile` как отдельные component types.

## Build engines (go family, local pilot)

| GP | `build.engine` | Sample repo | Jenkins job |
|----|----------------|-------------|-------------|
| `go-app` | `buildkit` | `samples/demo-go-app` | `demo-go-app` |
| `go-app-bp` | `buildpack` | `samples/demo-go-app-bp` | `demo-go-app-bp` |
| `go-app-df` | `dockerfile` | `samples/demo-go-app-df` | `demo-go-app-df` |

Content SoT:

```
coin-gp-content/stacks/
├── go-app/content.yaml       # buildkit
├── go-app-bp/content.yaml    # buildpack
└── go-app-df/content.yaml    # dockerfile
```

## Runtime pod

Один container — `manifest.runtime.image` (`coin-agent`), не отдельный stack agent.

См. [agent-build-model.md](agent-build-model.md).

## Pipeline stages

Typed stages в `content.yaml` — **без** script URLs:

```yaml
pipeline:
  stages:
    - id: validate
      name: Validate
    - id: test
      name: Test
    - id: build
      name: Build
    - id: publish
      name: Publish
      when: tag
```

Orchestration — `coin-lib` + `coin-executor`, не Groovy/shell из Nexus.

## Seed и publish (local)

```bash
cd docker
make coin-gp-content              # Gitea repo + job
make publish-agent                # coin-agent → Nexus
make seed-jenkins-lib             # components + GP profiles
make samples                      # product repos
make e2e-build-engines            # acceptance 3/3
```

How-to: [publish-gp-release.md](how-to/publish-gp-release.md).

## Catalog (local pilot)

| GP | Typical pin | Notes |
|----|-------------|-------|
| `go-app` | `1.0.2` | buildkit + Containerfile fixes |
| `go-app-bp` | `1.0.0` | buildpack |
| `go-app-df` | `1.0.2` | dockerfile targets |

Продукты могут pin `1.0.0` если GP release опубликован; после content bump — новый GP semver.

## Связанные документы

- [golden-path-versioning.md](golden-path-versioning.md)
- [config.md](config.md)
- [control-plane.md](control-plane.md)
