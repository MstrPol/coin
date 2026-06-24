# Design: Build Engine Model (completed)

ADR: [docs/adr/build-engine-contract.md](../../docs/adr/build-engine-contract.md)

## Hard cut decisions

| Removed | Replaced with |
|---------|---------------|
| coin-jenkins-agents stack images | coin-agent single image |
| GP scripts/*.sh runtime | coin-executor typed stages |
| dual-container pod (jnlp + stack) | single-container pod |
| manifest.jnlp | runtime.image from agent slot |
| curl bootstrap executor | executor baked in agent image |

## build.engine contract

```yaml
build:
  engine: buildkit | buildpack | dockerfile
```

Executor dispatches validate/test/build/publish per engine.

## Layer roles

| Layer | Role |
|-------|------|
| coin-gp-content | build policy, Containerfile/buildpack assets |
| coin-api | resolve → build, runtime.image, typed stages |
| coin-executor | engines + stages |
| coin-lib | glue only |

## E2E acceptance

`make e2e-build-engines` — buildkit, buildpack, dockerfile samples green.
