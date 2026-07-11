# build-engine

## ADDED Requirements

### Requirement: Build engine in manifest

Manifest SHALL include `build.engine` with value `buildkit`, `buildpack`, or `dockerfile`.

#### Scenario: Resolve emits build engine

- **WHEN** product resolves GP with gp-content defining build.engine
- **THEN** manifest MUST contain matching `build.engine` field

### Requirement: Single agent runtime image

Jenkins pod SHALL use single container from manifest `runtime.image` (coin-agent).

#### Scenario: No dual container pod

- **WHEN** coin-lib prepares K8s pod template
- **THEN** pod MUST NOT include separate jnlp and stack containers

### Requirement: Typed pipeline stages

Pipeline stages SHALL be typed executor stage names without shell script references.

#### Scenario: Stage execution via executor

- **WHEN** Jenkins runs Test stage
- **THEN** coin-lib MUST invoke coin-executor run --stage test without GP shell scripts

### Requirement: Build engine E2E

Local pilot SHALL pass E2E for all three engines.

#### Scenario: e2e-build-engines

- **WHEN** running make e2e-build-engines
- **THEN** buildkit, buildpack, and dockerfile sample jobs MUST succeed
