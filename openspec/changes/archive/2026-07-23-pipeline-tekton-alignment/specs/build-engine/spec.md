## MODIFIED Requirements

### Requirement: Typed pipeline stages

Pipeline tasks SHALL map to Jenkins stages by semantic `task.id`. Each task executes ordered v4 steps (`coin`, `containerfile`, `sh`) without GP shell scripts or Jenkins Shared Library business logic. coin-executor SHALL be invoked with `--task <id>` matching manifest task id.

#### Scenario: Task execution via v4 steps

- **WHEN** Jenkins runs task `test` from manifest with `kind: containerfile` step
- **THEN** coin-executor MUST dispatch the step using catalog ref and run config
- **AND** coin-lib MUST invoke executor without interpreting build logic in Groovy

### Requirement: Inline step dispatch in executor

coin-executor SHALL execute pipeline v4 steps directly from manifest `pipeline.tasks` and top-level `containerfiles` catalog without requiring separate manifest `build.targets` or `deliverables` sections.

#### Scenario: Execute containerfile step

- **WHEN** executor runs task step with `kind: containerfile`, `ref: app`, and catalog entry `app`
- **THEN** executor MUST materialize or open containerfile at catalog `path` and run the configured target
- **AND** MUST NOT require inline `containerfile.body` on the step

#### Scenario: Execute coin build then publish

- **WHEN** task `build-app` contains `kind: coin` build step and task `publish-app` contains publish step with `buildTaskId: build-app`
- **THEN** executor MUST merge build output during `build-app`
- **AND** publish step MUST publish the output entry named `build-app`

## ADDED Requirements

### Requirement: v4 step kind dispatch

coin-executor SHALL dispatch steps by `kind`: `coin` actions (`validate`, `test`, `build`, `publish`), `containerfile` catalog invoke, and allowlisted `sh` scripts.

#### Scenario: Coin validate step

- **WHEN** executor runs step `kind: coin` with `action: validate`
- **THEN** executor MUST run validate primitive without containerfile materialization

#### Scenario: Sh step allowlist enforcement

- **WHEN** executor runs step `kind: sh` not on allowlist
- **THEN** executor MUST fail before executing shell

### Requirement: Task flag for executor

coin-executor SHALL accept `--task <id>` to select `pipeline.tasks[].id`. Legacy `--stage` MUST remain as deprecated alias for Phase A.

#### Scenario: Run by task id

- **WHEN** Jenkins invokes `coin-executor run --manifest .coin/manifest.json --task validate`
- **THEN** executor MUST execute steps of task `validate` only

### Requirement: Build uses catalog id and computed image ref

For `kind: coin` with `action: build` and `type: image`, step MUST supply `containerfileRef` equal to a `containerfiles[].id` and `destinationRef` equal to a `destinations[].id`. coin-executor MUST compute image repository and tag from that destination's `imageRegistryPrefix`, product `.coin/config.yaml` `project.*`, and branching/version tag resolution. Optional step `cache` MUST override the destination's `buildCacheEnabled` when present. Step MUST NOT require a literal `imageRef` for the default GP contract.

#### Scenario: Build resolves containerfile by catalog id

- **WHEN** build step has `containerfileRef: app` and catalog entry `app` exists
- **THEN** executor MUST materialize or open that entry's `path` and build the image
- **AND** MUST NOT require `imageRef` in the step inputs

#### Scenario: Computed image ref from destinationRef

- **WHEN** build has `destinationRef: nexus-docker` with prefix `localhost:8082/coin-docker` and project `groupId=com.example.team`, `artifactId=demo`, `name=demo`
- **THEN** image repository path MUST be `localhost:8082/coin-docker/com.example.team/demo/demo` with tag from branching or env fallback

#### Scenario: Cache override on build step

- **WHEN** destination has `buildCacheEnabled: true` and build step sets `cache: false`
- **THEN** executor MUST NOT pass a cache ref to the build engine for that step

### Requirement: Build merges entry into outputs.json

Each successful image build task SHALL merge one entry into workspace `.coin/outputs.json`. Entry `name` MUST equal the build task `id`. Multiple build tasks MUST accumulate entries in the same file (merge by name), not create separate output files per build.

#### Scenario: Two builds write two output entries

- **WHEN** tasks `build-app` and `build-liquibase` each complete an image build
- **THEN** `.coin/outputs.json` MUST contain entries named `build-app` and `build-liquibase`
- **AND** each entry MUST include `type` and `ref`

### Requirement: Publish selects build output by buildTaskId

For `kind: coin` with `action: publish`, step MUST supply `buildTaskId` referencing a prior build task id and non-empty `destinationRefs[]` of `destinations[].id` entries with `push: true`. coin-executor MUST load `.coin/outputs.json`, select the entry whose `name` equals `buildTaskId`, and push that entry's image to each listed destination (retag when the destination prefix differs from the build ref). Publish MUST NOT require `containerfileRef` or a literal `imageRef` in step inputs. Publish MUST honor branching publish policy and Jenkins publish request semantics.

#### Scenario: Publish uses outputs from named build task

- **WHEN** publish step has `buildTaskId: build-app`, `destinationRefs: [nexus-docker]`, and outputs contain entry `build-app` with an image `ref`
- **THEN** executor MUST push that image to destination `nexus-docker`
- **AND** MUST NOT read `containerfiles` for the publish step

#### Scenario: Multi-push destinationRefs

- **WHEN** publish step lists two destination ids with `push: true`
- **THEN** executor MUST push (with retag if needed) to each destination in order

#### Scenario: Multiple publish tasks

- **WHEN** pipeline has publish tasks with `buildTaskId: build-app` and `buildTaskId: build-liquibase`
- **THEN** each publish MUST select and push its corresponding outputs entry

#### Scenario: Missing build output fails publish

- **WHEN** publish step references `buildTaskId: missing` with no matching outputs entry
- **THEN** executor MUST fail before push

### Requirement: Test runs in containerfile target without agent toolchain

coin-executor SHALL run unit/integration tests for the default GP path by invoking the build engine against a Containerfile catalog entry target (default target name `test`). Language toolchains MUST NOT be required on the Jenkins agent image. Default Phase A MUST NOT execute product test commands as shell on the agent.

Authoring MAY use `kind: containerfile` with `ref` and `run.target`, or `kind: coin` with `action: test` and `containerfileRef` / `target` as equivalent sugar.

#### Scenario: Test via containerfile target

- **WHEN** task `test` references catalog id `app` with target `test`
- **THEN** executor MUST run `podman build` (or the active engine equivalent) with that Containerfile path and `--target test`
- **AND** MUST NOT invoke `go test` or language package managers on the agent host

#### Scenario: Reject agent-side test cmd as default path

- **WHEN** a step attempts to run product unit tests only via unconstrained agent shell without a containerfile target
- **THEN** that path MUST NOT be the documented or seeded default for go-app-style GPs

### Requirement: Test reports exported to workspace

After a test task that produces reports, coin-executor SHALL export report files into workspace directory `.coin/test-results/`. Containerfile conventions MAY use an export stage (for example `test-reports`) and engine local output, or an equivalent extract. Executor SHOULD best-effort export reports even when the test target fails, then fail the task.

#### Scenario: Reports land under .coin/test-results

- **WHEN** test Containerfile writes coverage or junit into an exportable stage and test task completes export
- **THEN** workspace MUST contain files under `.coin/test-results/`
- **AND** those files MUST be readable by Jenkins on the agent without language toolchains

#### Scenario: Best-effort export on test failure

- **WHEN** the test target fails after writing report files into the exportable layout
- **THEN** executor SHOULD still populate `.coin/test-results/` when export is possible
- **AND** MUST fail the test task
