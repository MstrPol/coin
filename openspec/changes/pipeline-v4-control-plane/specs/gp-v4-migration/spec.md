## ADDED Requirements

### Requirement: Migrate v3 draft to v4 on save

When a GP draft still uses schemaVersion 3 (`pipeline.stages`, inline containerfile), coin-api SHALL convert it to schemaVersion 4 on save: stages → tasks with semantic ids, inline Containerfile bodies → `containerfiles[]` catalog entries with materializable path/contentRef, and destinations into catalog form when applicable.

#### Scenario: Save upgrades stages to tasks

- **WHEN** user saves a v3 draft with `pipeline.stages`
- **THEN** stored body MUST become `schemaVersion: 4` with `pipeline.tasks`
- **AND** MUST NOT keep `pipeline.stages` as the authoring model

### Requirement: Temporary v3 read adapter

Until reseed completes, coin-api MAY read existing v3 release blobs and adapt them to v4 resolved shape at materialize time. New writes MUST be schemaVersion 4 only after migration cutover for that GP.

#### Scenario: Legacy release still resolves

- **WHEN** remote resolve requests a GP version still stored as v3
- **THEN** materialize MUST produce a v4-shaped resolved manifest for Jenkins
- **OR** fail with an explicit migration-required error if adapter is disabled
