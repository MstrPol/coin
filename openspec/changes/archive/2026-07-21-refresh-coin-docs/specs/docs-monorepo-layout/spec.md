## ADDED Requirements

### Requirement: Workspace layout document

The project SHALL maintain `docs/workspace-layout.md` as the canonical description of the integration workspace: sibling component repositories, the `coin/` meta repository role, local pilot tooling under `coin/docker/`, and pointers to corp prod split.

#### Scenario: Contributor finds where code lives

- **WHEN** a new contributor opens `docs/README.md` or `docs/architecture.md`
- **THEN** they MUST find a link to `docs/workspace-layout.md`
- **AND** that document MUST list at least: `coin-api`, `coin-executor`, `coin-lib`, `coin-ui`, and `coin/` (docs, openspec, docker, starters)

#### Scenario: Removed trees documented

- **WHEN** reader consults workspace layout for historical paths `coin-gp-content`, `coin-branching-models`, or `coin-jenkins-agents`
- **THEN** the document MUST mark them removed/superseded
- **AND** MUST point to replacement SoT (api seed, docker testdata, coin-agent image)

### Requirement: Corp monorepo split runbook alignment

`docs/runbooks/prod-repo-split.md` SHALL describe corp-gate extraction of sibling repos into production Gitea projects and MUST NOT contradict `docs/workspace-layout.md` on current local layout.

#### Scenario: Local pilot does not execute corp split

- **WHEN** operator follows local pilot onboarding
- **THEN** docs MUST state that corp repo split is deferred to corp gate
- **AND** MUST NOT require creating separate prod Gitea repos for local compose correctness

#### Scenario: Inventory of corp target repos

- **WHEN** platform lead prepares P4-03
- **THEN** prod-repo-split MUST list target repos `coin-api`, `coin-executor`, `coin-ui`, `coin-lib`, `coin-starters`
- **AND** MUST NOT list `coin-gp-content` or `coin-branching-models` as required prod extract targets
