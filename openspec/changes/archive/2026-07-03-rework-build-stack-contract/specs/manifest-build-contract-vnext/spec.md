## ADDED Requirements

### Requirement: Manifest build contract vNext

Resolved manifest SHALL include a self-sufficient Build Stack vNext contract with `parameters`, `build.targets`, `deliverables`, `pipeline.stages`, `artifacts.containerfiles`, and `destinations`.

#### Scenario: Resolve emits build contract

- **WHEN** product resolves a GP release pinned to Build Stack vNext
- **THEN** manifest MUST include all parameters, targets, deliverables, pipeline stages, Containerfile content refs and destinations needed by executor
- **AND** manifest MUST NOT require live coin-api or PostgreSQL access during CI execution

#### Scenario: Manifest excludes credentials

- **WHEN** manifest contains Build Stack vNext contract
- **THEN** it MUST NOT contain Jenkins credential IDs, secret values or top-level `credentials`

### Requirement: Manifest target references are valid

coin-api SHALL validate all manifest references during materialization.

#### Scenario: Validate references

- **WHEN** Build Stack vNext contains stages, deliverables and targets
- **THEN** manifest builder MUST verify that every `targetId`, `deliverableId`, parameter reference and Containerfile artifact reference resolves

#### Scenario: Reject invalid manifest

- **WHEN** Build Stack vNext contains unresolved reference
- **THEN** coin-api MUST reject draft preview/promote instead of emitting partial manifest

### Requirement: Executor uses manifest contract only

coin-executor SHALL execute build, test and publish stages from manifest Build Stack vNext contract and product identity.

#### Scenario: Executor runs stage from manifest

- **WHEN** Jenkins invokes `coin-executor run --stage build`
- **THEN** executor MUST read `pipeline.stages` from manifest
- **AND** execute referenced platform actions against targets and deliverables
- **AND** MUST NOT read build targets from product `.coin/config.yaml`

### Requirement: Manifest hash includes Build Stack vNext

Manifest integrity metadata SHALL change when Build Stack vNext parameters, targets, deliverables, artifacts or stages change.

#### Scenario: Hash changes on target edit

- **WHEN** publisher changes a target engine or Containerfile artifact reference
- **THEN** resolved manifest hash MUST change
