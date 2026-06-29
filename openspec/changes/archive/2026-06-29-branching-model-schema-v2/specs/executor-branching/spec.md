## MODIFIED Requirements

### Requirement: Branch validation at validate stage

coin-executor validate SHALL match the current git branch against ordered `manifest.branching.branches` rules (first match wins) and reject branches that match no rule or violate the matched pattern.

#### Scenario: Feature branch allowed

- **WHEN** current branch matches a feature rule pattern
- **THEN** validate MUST pass branch check

#### Scenario: Invalid branch name

- **WHEN** branch name matches no rule or fails the matched rule pattern
- **THEN** validate MUST fail with actionable error

### Requirement: Product version from git

coin-executor version SHALL compute product version from the matched branch rule `versioning.template`, named regex captures, and git tag history.

#### Scenario: Snapshot version on feature branch

- **WHEN** on a feature branch with template `v{base}-{jira}-snapshot-{n}`
- **THEN** COIN_VERSION MUST follow the template series for that branch

### Requirement: Publish policy enforcement

Publish stage eligibility SHALL be controlled by developer publish request plus per-branch `publish` flag on the matched rule. Automatic publish on tag alone MUST NOT occur.

#### Scenario: Publish not requested

- **WHEN** publish stage runs without publish request (`COIN_PUBLISH_REQUEST` not true)
- **THEN** executor MUST skip publish with exit 0 and clear message

#### Scenario: Publish denied on ineligible branch

- **WHEN** publish is requested and matched branch rule has `publish: false`
- **THEN** executor MUST fail publish with non-zero exit and actionable error

#### Scenario: Publish allowed on eligible branch

- **WHEN** publish is requested and matched branch rule has `publish: true`
- **THEN** executor MUST run publish stage

#### Scenario: Stage policy not bypassed

- **WHEN** Jenkins invokes `run --stage publish` with publish requested
- **THEN** executor MUST still evaluate branch eligibility (no bypass)
