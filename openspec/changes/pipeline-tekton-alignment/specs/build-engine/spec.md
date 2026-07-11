## MODIFIED Requirements

### Requirement: Typed pipeline stages

Pipeline tasks SHALL map to Jenkins stages by semantic `task.id`. Each task executes ordered v4 steps (`coin`, `containerfile`, `sh`) without GP shell scripts or Jenkins Shared Library business logic. coin-executor SHALL be invoked with `--task <id>` matching manifest task id.

#### Scenario: Task execution via v4 steps

- **WHEN** Jenkins runs task `test` from manifest with `kind: containerfile` step
- **THEN** coin-executor MUST dispatch the step using catalog ref and run config
- **AND** coin-lib MUST invoke executor without interpreting build logic in Groovy

### Requirement: Inline step dispatch in executor

coin-executor SHALL execute pipeline v4 steps directly from manifest `pipeline.tasks` and `containerfiles` catalog without requiring separate manifest `build.targets` or `deliverables` sections.

#### Scenario: Execute containerfile step

- **WHEN** executor runs task step with `kind: containerfile`, `ref: app`, and manifest catalog entry `app` with `contentRef`
- **THEN** executor MUST materialize containerfile from catalog ref and run the configured target
- **AND** MUST NOT require inline `containerfile.body` on the step

#### Scenario: Execute coin build then publish

- **WHEN** task `build` contains `kind: coin` build step and task `publish` contains publish step with `buildTaskId: build`
- **THEN** executor MUST build output during build task
- **AND** publish step MUST publish the output associated with task `build`

## ADDED Requirements

### Requirement: v4 step kind dispatch

coin-executor SHALL dispatch steps by `kind`: `coin` actions (`validate`, `test`, `build`, `publish`), `containerfile` catalog invoke, and allowlisted `sh` scripts.

#### Scenario: Coin validate step

- **WHEN** executor runs step `kind: coin` with `action: validate`
- **THEN** executor MUST run validate primitive without containerfile materialization

#### Scenario: Sh step allowlist enforcement

- **WHEN** executor runs step `kind: sh` not on allowlist
- **THEN** executor MUST fail before executing shell
