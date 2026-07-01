## Why

Resolved manifest currently contains fields that are not sourced from the Golden Path release composition, for example Jenkins credential IDs injected by `coin-api`. This makes GP releases environment-specific and blurs the boundary between platform content, product/Jenkins configuration, and resolve-time materialization.

This change tightens the manifest contract after the runtime and branching-model cleanup: `coin-api` resolve must only denormalize the selected GP release identity plus the three composition sections (`agent`, `gp-content`, `branching-model`).

## What Changes

- **BREAKING**: Remove `credentials` from resolved manifest output and schema.
- Require `coin-api` manifest builder to emit only:
  - GP release identity: `goldenPath.name`, `goldenPath.version`
  - `agent` materialization: `runtime.image`, `runtime.digest`
  - `gp-content` materialization: `build`, `pipeline`, `validateSchema`, `capabilities`
  - `branching-model` materialization: `branching`
  - manifest metadata required for cache/fallback integrity: `manifestVersion`, `manifestHash`
- Move Jenkins credential selection fully outside GP resolve:
  - product `.coin/config.yaml` / starter config
  - `coin-lib` defaults and environment/Jenkins configuration
- Update docs and tests that currently imply credential IDs are part of the manifest.

## Non-goals

- Do not change GP composition slots; the release remains three pins: `agent`, `gp-content`, `branching-model`.
- Do not add a new manifest version unless implementation discovers a compatibility requirement.
- Do not introduce credential registry, secret references, or per-environment bindings in `coin-api`.
- Do not change build engine behavior, branching policy evaluation, or publish gate semantics.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `gp-composition-two-slot`: resolved manifest section ownership and allowed materialized fields.
- `jenkins-lib-boundary`: Jenkins credential binding must remain in `coin-lib`/project configuration, not in `coin-api` resolve.

## Impact

- `coin-api/internal/manifest/builder.go` and manifest schema/tests.
- `coin-lib` config merge logic that currently reads `manifest.credentials`.
- Product starters and docs that describe `manifest.credentials` or use manifest-provided credential IDs.
- Nexus fallback manifest blobs after re-resolve/publish will no longer include Jenkins credential IDs.
