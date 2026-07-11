## MODIFIED Requirements

### Requirement: Pipeline-inline canonical model

Pipeline-inline SHALL use `schemaVersion: 4` where pipeline configuration lives in `pipeline.tasks[]` with typed steps (`coin`, `containerfile`, `sh`) and named `containerfiles[]` catalog on the GP release body. The structured model SHALL be stored as the embedded body of a GP release draft, not as a separate `gp-content` component package. Author-facing model MUST NOT include top-level `build.targets`, `deliverables`, or legacy v2 `artifacts.containerfiles`. Author-facing model MUST NOT use `pipeline.stages[]` for new v4 drafts.

#### Scenario: Valid v4 GP release pipeline shape

- **WHEN** publisher saves pipeline-inline v4 model on GP draft `go-app@1.0.0-snapshot.1`
- **THEN** coin-api MUST persist `parameters`, `validateSchema`, `containerfiles`, and `pipeline.tasks` on that GP release
- **AND** MUST NOT create a `gp-content` component version

#### Scenario: Reject v2 catalog sections in GP pipeline draft

- **WHEN** GP release pipeline body contains `schemaVersion: 4` together with `build.targets`, `deliverables`, or legacy `artifacts.containerfiles`
- **THEN** validate MUST reject the draft

### Requirement: Parameters unchanged from vNext

Pipeline-inline SHALL retain typed non-secret parameters with the same validation rules as Build Stack vNext.

#### Scenario: Enum parameter requires allowed values

- **WHEN** parameter has `type: enum`
- **THEN** validation MUST require non-empty `allowedValues`

## REMOVED Requirements

### Requirement: Typed inline pipeline steps

**Reason**: Superseded by v4 step kinds `coin`, `containerfile`, `sh` on `pipeline.tasks[]`.

**Migration**: Map v3 `action: run|build|publish` to v4 `coin` and `containerfile` steps per migration rules in `pipeline-tekton-model`.

### Requirement: Short hash stage and build ids

**Reason**: Replaced by semantic `task.id` and `publish.buildTaskId` references.

**Migration**: v3→v4 migration assigns semantic ids from `stage.name` or positional defaults; published v3 remains readable via adapter.

### Requirement: Containerfile inline in buildkit steps

**Reason**: Managed Containerfile content moves to `containerfiles[]` catalog; steps reference by `ref`.

**Migration**: Extract inline `containerfile.body` into catalog entries on save; materialize `contentRef` in manifest.

### Requirement: BYO dockerfile inline in step

**Reason**: BYO paths move to `containerfiles[]` catalog entries with `kind: project`.

**Migration**: Map `dockerfile.path` on step to catalog entry `kind: project` referenced by containerfile step.
