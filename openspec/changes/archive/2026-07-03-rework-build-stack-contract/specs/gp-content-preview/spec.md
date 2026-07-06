## MODIFIED Requirements

### Requirement: GP content preview API

coin-api SHALL expose `POST /v1/admin/gp-content/preview` accepting Build Stack vNext canonical model and returning resolved manifest preview plus validation issues and engine-specific warnings.

#### Scenario: Preview buildkit target

- **WHEN** publisher posts valid Build Stack vNext model with buildkit target and managed Containerfile artifact
- **THEN** coin-api MUST return resolved `build.targets[]` entry for that target
- **AND** MUST include Containerfile content ref in `artifacts.containerfiles`

#### Scenario: Preview BYO dockerfile target

- **WHEN** publisher posts valid Build Stack vNext model with dockerfile target and workspace Dockerfile path
- **THEN** coin-api MUST return resolved `build.targets[]` entry without managed Containerfile content ref for that target

#### Scenario: Preview parameters and deliverables

- **WHEN** publisher posts Build Stack vNext model with parameters and deliverables
- **THEN** coin-api MUST return resolved manifest preview containing `parameters` and structured `deliverables`
- **AND** MUST validate references from stages and deliverables to targets

#### Scenario: Reject legacy v2 content only body

- **WHEN** publisher posts legacy `content.yaml` v2 body without Build Stack vNext model
- **THEN** preview MUST return validation error for Build Stack vNext endpoints
