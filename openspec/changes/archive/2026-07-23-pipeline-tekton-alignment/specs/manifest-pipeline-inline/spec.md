## MODIFIED Requirements

### Requirement: Pipeline-inline resolved manifest

Resolved manifest SHALL contain `parameters`, `validateSchema` ref when applicable, `containerfiles[]` for managed/project entries, and `pipeline.tasks` with typed steps. Containerfile steps MUST resolve catalog refs to catalog entries. Manifest MUST NOT contain top-level `build.targets`, `deliverables`, legacy `artifacts.containerfiles`, or `capabilities.deliverables` for schemaVersion 4. Manifest MUST NOT use `pipeline.stages` for schemaVersion 4.

Resolved manifest MAY be obtained from coin-api/Nexus **or** from product file resolve (`coin.resolve: file`).

#### Scenario: File resolve provides executable v4 manifest

- **WHEN** Jenkins resolves with `coin.resolve: file` and a valid v4 fixture
- **THEN** materialized `.coin/manifest.json` MUST include `pipeline.tasks`
- **AND** coin-executor MUST run tasks without calling coin-api

### Requirement: Manifest self-sufficiency for offline execution

Pipeline-inline manifest SHALL be sufficient for coin-executor without live PostgreSQL. Executor MUST obtain Containerfile materialization from manifest `containerfiles` catalog and step refs (or project paths).

#### Scenario: Executor runs from fixture only

- **WHEN** workspace has `.coin/manifest.json` from file resolve and no coin-api
- **THEN** executor MUST execute `coin` and `containerfile` steps from that manifest
- **AND** MUST NOT require live GP draft lookup

### Requirement: Publish destinations compatibility

For schemaVersion 4, top-level `destinations` MUST be a catalog array of named registry entries (`id`, `imageRegistryPrefix`, optional `host`/`pull`/`push`/`buildCacheEnabled`/`artifactRepositoryBase`). Build steps MUST reference a destination via `destinationRef`; publish steps MUST use non-empty `destinationRefs[]` for multi-push. Destinations supply registry prefix for build-time image ref computation; publish does not re-author `imageRef` in the step. coin-lib MUST authenticate to every destination host with `pull: true` or `push: true` before stages that need registry access.

#### Scenario: Publish image step from outputs

- **WHEN** pipeline includes publish step with `buildTaskId` and `destinationRefs` for a prior image build task
- **THEN** executor MUST push the `ref` stored under that build task name in `.coin/outputs.json` to each listed destination

#### Scenario: Local pilot single destination

- **WHEN** demo fixture declares one destination `nexus-docker` with `pull: true` and `push: true`
- **THEN** build MAY use `destinationRef: nexus-docker` and publish MAY use `destinationRefs: [nexus-docker]`
- **AND** registry auth MUST cover that destination host

#### Scenario: Top-level containerfiles in resolved manifest

- **WHEN** v4 resolved manifest is materialized or loaded from file
- **THEN** it MUST expose `containerfiles[]` at the top level with `id`, `kind`, and `path` per entry
- **AND** managed entries MUST be sufficient to obtain content at `path` via `contentRef` fetch or an on-disk file
- **AND** MUST NOT rely on inline `body` in the resolved/file document
