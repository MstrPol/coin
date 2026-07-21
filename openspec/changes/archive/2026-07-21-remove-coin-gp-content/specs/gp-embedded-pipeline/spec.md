## ADDED Requirements

### Requirement: Bootstrap seed without coin-gp-content tree

Local pilot bootstrap and GP release default pipeline bodies SHALL be sourced from embedded seed under `coin-api/internal/gpcontent/seed/` (or equivalent embed in coin-api). The workspace tree `coin/coin-gp-content/` MUST NOT be required for authoring, resolve, seed, or publish of GP releases.

#### Scenario: Seed GP release without coin-gp-content folder

- **WHEN** operator runs local seed that creates GP profile/release with embedded pipeline (`go-app` or `go-app-docker`)
- **THEN** pipeline defaults MUST come from coin-api seed embed
- **AND** MUST NOT read files from `coin/coin-gp-content/`

#### Scenario: No local Jenkins publish job for coin-gp-content

- **WHEN** operator bootstraps local docker stack for Coin platform
- **THEN** bootstrap MUST NOT require `make coin-gp-content` or a Jenkins job that mirrors `coin-gp-content/` into Gitea for pilot correctness
