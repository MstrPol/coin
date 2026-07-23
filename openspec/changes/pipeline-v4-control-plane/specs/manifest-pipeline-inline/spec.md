## ADDED Requirements

### Requirement: Remote resolve emits v4 fixture-equivalent shape

When coin-api materializes a schemaVersion 4 GP for remote resolve, the resolved manifest served to Jenkins SHALL be equivalent in structure to the file-resolve fixture contract: `pipeline.tasks`, top-level `containerfiles[]`, and `destinations[]` catalog. coin-lib and coin-executor MUST NOT require a separate adapter for remote vs file shapes.

#### Scenario: Remote resolve without file fixture

- **WHEN** product config uses `coin.resolve: remote` (default) against a seeded v4 GP
- **THEN** materialized `.coin/manifest.json` MUST include `pipeline.tasks` and `containerfiles`
- **AND** coin-executor MUST run validate/test/build without `coin.resolve: file`
