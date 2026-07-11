# executor-branching

## ADDED Requirements

### Requirement: Branch validation at validate stage

coin-executor validate SHALL reject builds on branches that violate `manifest.branching` rules.

#### Scenario: Feature branch allowed

- **WHEN** current branch matches trunk-based feature pattern
- **THEN** validate MUST pass branch check

#### Scenario: Invalid branch name

- **WHEN** branch name violates naming pattern
- **THEN** validate MUST fail with actionable error

### Requirement: Product version from git

coin-executor version SHALL output product version computed from branching rules and git state, not executor binary version.

#### Scenario: Snapshot version on feature branch

- **WHEN** on feature branch with trunk-based model
- **THEN** COIN_VERSION MUST follow model versioning format

### Requirement: Publish policy enforcement

Publish stage SHALL run only when branching publish policy allows (e.g. rc tag on release branch).

#### Scenario: Publish skipped on feature branch

- **WHEN** publish stage runs on feature branch without matching tag
- **THEN** executor MUST skip publish with exit 0 and clear message

#### Scenario: Stage policy not bypassed

- **WHEN** Jenkins invokes `run --stage publish` on ineligible branch
- **THEN** executor MUST still evaluate branching policy (no bypass)

### Requirement: Image tag uses COIN_VERSION

Docker image tags in publish SHALL use COIN_VERSION from ResolveVersion.

#### Scenario: Publish tags image

- **WHEN** publish succeeds on eligible tag
- **THEN** image ref MUST use computed COIN_VERSION not GP pin version
