## Context

After the runtime and branching-model cleanup, GP release composition has exactly three operator pins: `agent`, `gp-content`, and `branching-model`. The resolved manifest should be a deterministic materialization of those pins plus the GP release identity.

Today `coin-api` adds `credentials: { docker: "nexus-docker" }` in the manifest builder. That value is a Jenkins/site credential ID, not platform content. It makes the manifest environment-specific and lets `coin-api` influence Jenkins credential binding even though `coin-lib` is the only allowed Jenkins glue layer.

## Goals / Non-Goals

**Goals:**

- Make resolved manifest ownership explicit: GP identity plus sections sourced from `agent`, `gp-content`, and `branching-model`.
- Remove Jenkins credential IDs from `coin-api` resolve output and manifest schema.
- Ensure `coin-lib` chooses credential IDs from product config, library defaults, or Jenkins environment configuration.
- Keep Nexus fallback valid by preserving manifest metadata (`manifestVersion`, `manifestHash`) and immutable content refs.

**Non-Goals:**

- No change to the three-pin GP composition model.
- No credential registry or secret-reference abstraction in `coin-api`.
- No change to build engine dispatch, branching policy semantics, or publish gates.
- No corp fleet rollout work in this change.

## Decisions

### D1. Resolve is a pure materializer of GP release composition

`coin-api` SHALL emit only the fields that come from the resolved GP release identity and the three component pins:

| Source | Manifest fields |
|--------|-----------------|
| GP release | `goldenPath.name`, `goldenPath.version` |
| agent | `runtime.image`, `runtime.digest` |
| gp-content | `build`, `pipeline`, `validateSchema`, `capabilities` |
| branching-model | `branching` |
| resolve integrity | `manifestVersion`, `manifestHash` |

Alternative considered: keep `credentials` as a convenience default. Rejected because credential IDs are Jenkins-instance local and violate the coin-lib boundary.

### D2. Jenkins credential selection stays in coin-lib/project config

`coin-lib` already reads product `.coin/config.yaml` and library defaults. The manifest layer must not override `jenkins.credentials.*`. Implementation should remove the manifest-to-config merge for `manifest.credentials` and rely on:

1. Product config (`jenkins.credentials.docker`)
2. `coin-lib` defaults (`coin-lib-defaults.yaml`)
3. Environment/Jenkins configuration if supported by existing glue

Alternative considered: expose credential aliases in GP content. Rejected because aliases still encode Jenkins binding policy in GP content and would need a separate security model.

### D3. Hard cut without compatibility shim

This is local pilot behavior, not a shipped corp contract. Remove `credentials` from manifest schema and tests instead of retaining deprecated compatibility fields.

Alternative considered: keep schema allowing `credentials` while builder stops emitting it. Rejected because it leaves an attractive but invalid contract in docs and generated examples.

## Risks / Trade-offs

- Existing local Jenkins jobs may have relied on manifest-provided `nexus-docker` → Ensure product starter configs and `coin-lib` defaults still provide the same value.
- Cached Nexus manifest blobs may still contain `credentials` until re-resolved → Re-run resolve/publish for affected pilot GPs after implementation.
- Some docs/tests may describe credential IDs as manifest fields → Update docs and schema tests in the same change.

## Migration Plan

1. Update OpenSpec/doc contract.
2. Remove `credentials` emission from `coin-api` manifest builder and schema.
3. Remove `manifest.credentials` merge from `coin-lib`.
4. Confirm product starter/sample configs still contain `jenkins.credentials.docker` or defaults cover local pilot.
5. Re-resolve `gp-01-07@1.0.0` and verify manifest has no `credentials` key.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Нужен ли manifestVersion bump при удалении `credentials`? | ✅ | A: bump to v2; B: hard cut v1 | B: hard cut v1 для local pilot, потому что поле не должно было быть частью контракта |
