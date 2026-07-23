## ADDED Requirements

### Requirement: Seed schemaVersion 4 pipelines

Platform seed for `go-app` and `go-app-docker` SHALL publish GP release bodies with `schemaVersion: 4` (`pipeline.tasks` + `containerfiles` catalog) so remote resolve E2E does not depend on product `resolve: file`.

#### Scenario: Reseed enables remote demo-go-app

- **WHEN** operators run reseed after v4 seed update
- **THEN** remote resolve for demo-go-app MUST obtain a v4 pipeline body
- **AND** Jenkins build MUST succeed without a local `manifest.local.yaml` as primary resolve path
