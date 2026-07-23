## ADDED Requirements

### Requirement: Top-level containerfiles catalog

Resolved and authoring v4 documents SHALL include a top-level `containerfiles[]` section as a peer to `branching`, `parameters`, and runtime/agent sections. Pipeline steps MUST reference catalog entries by id only. Managed Containerfile body MUST NOT appear inline on pipeline steps.

#### Scenario: Peer section alongside branching and parameters

- **WHEN** a v4 resolved fixture or GP pipeline body is validated
- **THEN** `containerfiles` MUST be a top-level array when any step references a containerfile
- **AND** MUST NOT be nested under `pipeline.tasks[].steps`

#### Scenario: Multiple catalog entries

- **WHEN** document defines `containerfiles` with ids `app` and `liquibase`
- **THEN** build or containerfile steps MUST reference entries by id only
- **AND** executor MUST resolve each id through the catalog

### Requirement: Catalog entry fields id kind path

Each `containerfiles[]` entry SHALL declare unique `id`, `kind` of `managed` or `project`, and workspace-relative `path`. Duplicate ids or two entries sharing the same `path` MUST be rejected. Default path for managed entries without explicit `path` SHALL be `.coin/containerfiles/<id>`.

#### Scenario: Accept managed and project entries

- **WHEN** catalog contains `kind: managed` with `path: .coin/containerfiles/app` and `kind: project` with `path: Dockerfile`
- **THEN** validation MUST accept both entries

#### Scenario: Reject duplicate catalog id

- **WHEN** two catalog entries share the same `id`
- **THEN** validation MUST reject the document

#### Scenario: Reject duplicate path

- **WHEN** two catalog entries declare the same `path`
- **THEN** validation MUST reject the document

#### Scenario: Default path for managed without explicit path

- **WHEN** managed entry omits `path`
- **THEN** materialize or validate MUST use default `.coin/containerfiles/<id>`

### Requirement: Containerfile content lives on disk under path

Workspace Containerfile content for catalog entries SHALL live in files at `containerfiles[].path` (conventionally under `.coin/containerfiles/`). Resolved manifests and `coin.resolve: file` fixtures MUST declare `id`, `kind`, and `path` and MUST NOT embed full Containerfile text in `containerfiles[].body`.

#### Scenario: File-resolve fixture without inline body

- **WHEN** product uses `coin.resolve: file` with catalog entry `id: app`, `kind: managed`, `path: .coin/containerfiles/app`
- **THEN** the fixture YAML/JSON MUST NOT include `body` for that entry
- **AND** workspace MUST contain the Containerfile file at `.coin/containerfiles/app`

#### Scenario: Resolved remote catalog uses path not body

- **WHEN** coin-api materializes a managed catalog entry into a resolved manifest
- **THEN** the manifest entry MUST include `path` (default `.coin/containerfiles/<id>`) and MAY include `contentRef`/`digest`
- **AND** MUST NOT require consumers to read Containerfile text from a `body` field on the resolved document

### Requirement: Managed containerfile materialize to path

Catalog entry with `kind: managed` SHALL obtain content from authoring `body` (GP draft only), or resolved `contentRef`/`digest`, or an already-present file at `path` (file-resolve / pre-materialized workspace). Before build or containerfile invoke, coin-executor SHALL ensure the file exists at entry `path`.

#### Scenario: Executor fetches managed content to path

- **WHEN** task step references catalog id `app` with `kind: managed`, `path: .coin/containerfiles/app`, and `contentRef` is set
- **THEN** executor MUST write fetched content to that path
- **AND** MUST run the build engine against that path
- **AND** MUST NOT require inline `containerfile.body` on the step

#### Scenario: Executor uses pre-seeded managed path for file resolve

- **WHEN** managed entry has `path: .coin/containerfiles/app`, no `body`, no `contentRef`, and the file already exists
- **THEN** executor MUST use the existing file at that path
- **AND** MUST NOT fail solely because `body` is absent

#### Scenario: Managed missing content and missing file fails

- **WHEN** managed entry has no `body`, no `contentRef`, and `path` file is missing
- **THEN** executor MUST fail with an error naming the catalog id and path

### Requirement: Project containerfile uses workspace path

Catalog entry with `kind: project` SHALL include workspace-relative `path`. Executor MUST use the existing workspace file and MUST NOT fetch managed body from Nexus for that entry.

#### Scenario: BYO project path entry

- **WHEN** catalog entry has `kind: project` and `path: Dockerfile`
- **THEN** validation MUST accept entry without `body` or `contentRef`
- **AND** executor MUST use workspace file at `path` when a step references the entry

### Requirement: Catalog entry id format

`containerfiles[].id` SHALL match `^[a-z][a-z0-9-]{1,31}$`.

#### Scenario: Accept semantic catalog id

- **WHEN** catalog `id` is `app` or `liquibase-db`
- **THEN** validation MUST accept the id
