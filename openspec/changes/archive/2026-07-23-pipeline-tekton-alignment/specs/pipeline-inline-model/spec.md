## MODIFIED Requirements

### Requirement: Pipeline-inline canonical model

Pipeline-inline SHALL use `schemaVersion: 4` where pipeline configuration lives in `pipeline.tasks[]` with typed steps (`coin`, `containerfile`, `sh`) and a top-level `containerfiles[]` catalog (`id`, `kind: managed|project`, `path`). The structured model is the embedded body of a GP release (Phase B storage) and the corresponding sections of a resolved manifest / file fixture (Phase A). Author-facing model MUST NOT include top-level `build.targets`, `deliverables`, legacy v2 `artifacts.containerfiles`, or `capabilities.deliverables`. Author-facing model MUST NOT use `pipeline.stages[]` for new v4 documents. Build steps reference catalog ids only; publish steps reference build task ids (see `build-engine`).

#### Scenario: Valid v4 document shape

- **WHEN** a v4 pipeline-inline or resolved fixture is validated
- **THEN** it MUST contain `pipeline.tasks` and MAY contain `containerfiles`, `parameters`, `validateSchema`
- **AND** MUST NOT require a `gp-content` component package to explain pipeline content
- **AND** MUST NOT require `capabilities.deliverables`

#### Scenario: Reject v2 catalog sections in v4 document

- **WHEN** document has `schemaVersion: 4` together with `build.targets`, `deliverables`, or legacy `artifacts.containerfiles`
- **THEN** validation MUST reject the document

### Requirement: Parameters unchanged from vNext

Pipeline-inline SHALL retain typed non-secret parameters with the same validation rules as Build Stack vNext.

#### Scenario: Enum parameter requires allowed values

- **WHEN** parameter has `type: enum`
- **THEN** validation MUST require non-empty `allowedValues`

## REMOVED Requirements

### Requirement: capabilities.deliverables as GP build capability list

**Reason**: schemaVersion 4 declares build outputs via `pipeline.tasks` (`kind: coin` + `action: build` + `build.type`) and `.coin/outputs.json`.

**Migration**: omit `capabilities` / `capabilities.deliverables` from v4 fixtures and resolved manifests.

### Requirement: Typed inline pipeline steps

**Reason**: Superseded by v4 step kinds `coin`, `containerfile`, `sh` on `pipeline.tasks[]`.

**Migration**: Map v3 `action: run|build|publish` to v4 `coin` and `containerfile` steps per migration rules in `pipeline-tekton-model`.

### Requirement: Short hash stage and build ids

**Reason**: Replaced by semantic `task.id` and `publish.buildTaskId` references.

**Migration**: Phase B v3→v4 migration assigns semantic ids; Phase A fixtures use semantic ids directly.

### Requirement: Containerfile inline in buildkit steps

**Reason**: Managed Containerfile content moves to `containerfiles[]` catalog; steps reference by `ref`.

**Migration**: Extract inline `containerfile.body` into catalog entries; materialize `contentRef` in remote manifests (Phase B).

### Requirement: BYO dockerfile inline in step

**Reason**: BYO paths move to `containerfiles[]` catalog entries with `kind: project`.

**Migration**: Map `dockerfile.path` on step to catalog entry `kind: project` referenced by containerfile step.
