## Why

До hard cut build-модель Coin использовала language-specific stack agents, GP shell scripts, dual-container Jenkins pods и bootstrap curl executor — противоречиво и тяжело в сопровождении.

Нужен единый контракт: `build.engine` в GP/manifest, typed executor stages, один `coin-agent` image.

**Статус:** реализовано (local pilot E2E 3/3). Change для archive → baseline specs.

## What Changes

- `build.engine`: buildkit | buildpack | dockerfile в gp-content/manifest
- Единый `coin-executor/Dockerfile.agent` (inbound-agent + executor baked in)
- Typed `pipeline.stages` без script refs
- Удалены coin-jenkins-agents, manifest.jnlp, dual-container pod
- **BREAKING:** hard cut без dual path

## Capabilities

### New Capabilities

- `build-engine`: engine dispatch, manifest build object, coin-agent runtime

### Modified Capabilities

- _(baseline sync при archive)_

## Impact

- coin-gp-content, coin-api, coin-executor, coin-lib
- docs/agent-build-model.md, ADR [build-engine-contract](../../docs/adr/build-engine-contract.md)

## Non-goals

- Project-level build.engine override
- Legacy script URLs в manifest
