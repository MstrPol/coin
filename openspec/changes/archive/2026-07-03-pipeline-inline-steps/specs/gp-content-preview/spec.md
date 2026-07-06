## MODIFIED Requirements

### Requirement: GP content preview API

coin-api SHALL expose `POST /v1/admin/gp-content/preview` accepting pipeline-inline `schemaVersion: 3` structured model and returning resolved manifest preview, validation issues and warnings.

#### Scenario: Preview pipeline-inline draft

- **WHEN** publisher posts valid v3 pipeline-inline structured model
- **THEN** coin-api MUST return resolved manifest with inline pipeline stages and per-step containerfile content refs
- **AND** MUST NOT require separate targets, deliverables or containerfile catalogs in request body

#### Scenario: Preview validation for inline publish ref

- **WHEN** publisher posts v3 model where publish step references missing `buildStepId`
- **THEN** preview MUST return validation error with field path to the publish step

#### Scenario: Reject v2 catalog-only draft

- **WHEN** publisher posts v2 model with `build.targets` and without v3 pipeline-inline shape
- **THEN** preview MUST return validation error indicating schemaVersion 3 is required

## REMOVED Requirements

### Requirement: Preview buildkit draft

**Reason**: v2 `content.yaml` engine-centric preview replaced by pipeline-inline structured model preview.
**Migration**: Use structured v3 model POST body; preview returns inline pipeline manifest.

### Requirement: Preview BYO dockerfile draft

**Reason**: BYO dockerfile validation moves to inline run/build steps with `engine: dockerfile`.
**Migration**: Post v3 model with dockerfile fields inside inline steps.

### Requirement: Reject v1 content

**Reason**: Hard cut already on v2+; v3 continues hard cut without v1 support.
**Migration**: Upgrade stacks to schemaVersion 3.
