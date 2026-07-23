# jenkins-lib-boundary Specification

## Purpose

Boundary between Jenkins Shared Library (`coin-lib`) glue and control plane: lib outside coin-api registry; resolve, DAG expansion, file resolve, test archive, registry auth.

## Requirements

### Requirement: Jenkins lib outside control plane

The coin-api control plane SHALL NOT register, store, or expose Jenkins Shared Library (`coin-lib`) as a platform component.

Jenkins glue version SHALL be managed outside coin-api (Jenkins org configuration, HTTP retriever, Nexus immutable ZIP).

#### Scenario: No lib component type in registry

- **WHEN** client lists platform components or queries admin component API
- **THEN** coin-api MUST NOT return components with type `lib`

#### Scenario: No lib admin registration API

- **WHEN** client POSTs to register a lib component version
- **THEN** coin-api MUST respond with HTTP 404 Not Found (route removed)

### Requirement: coin-lib Nexus-only publish

The coin-lib repository publish path SHALL upload immutable ZIP artifacts to Nexus only and SHALL NOT call coin-api admin registration.

#### Scenario: Publish without coin-api

- **WHEN** `publish-lib.sh` completes successfully for version `1.0.0`
- **THEN** Nexus MUST contain the ZIP at the configured maven-releases path
- **AND** coin-api PostgreSQL MUST NOT contain a `lib` component row for that version

### Requirement: Manifest resolve excludes lib

Resolved GP manifest SHALL NOT include a `lib` section.

coin-api SHALL NOT read or validate a platform lib pin during resolve.

#### Scenario: Resolve manifest shape

- **WHEN** client resolves manifest for GP `go-app@1.0.0`
- **THEN** the JSON document MUST NOT contain top-level key `lib`
- **AND** MUST contain runtime/executor materialization from agent stack and gp-content/branching pins

### Requirement: Product Jenkins bootstrap independent of coin-api lib API

Product CI SHALL load coin-lib via Jenkins Shared Library configuration (`@Library`) without calling coin-api for lib version selection.

#### Scenario: No LibraryVersion endpoint

- **WHEN** client calls `GET /v1/golden-paths/{name}/version`
- **THEN** coin-api MUST respond with HTTP 404 Not Found

### Requirement: Jenkins credentials outside resolved manifest

Jenkins credential IDs SHALL be selected by product/Jenkins configuration and `coin-lib` defaults, not by `coin-api` resolved manifest.

Resolved manifest SHALL NOT contain a top-level `credentials` section or any Jenkins-instance credential ID such as `nexus-docker`.

`coin-lib` SHALL NOT merge `manifest.credentials` into the effective project configuration. Docker registry credential binding SHALL use `jenkins.credentials.docker` from product config or the existing `coin-lib` default chain.

#### Scenario: Resolve manifest excludes Jenkins credential IDs

- **WHEN** product CI resolves a GP manifest
- **THEN** the returned JSON document MUST NOT contain top-level key `credentials`
- **AND** MUST NOT contain Jenkins credential ID values used for local registry binding

#### Scenario: coin-lib binds Docker credentials from project config

- **WHEN** product `.coin/config.yaml` contains `jenkins.credentials.docker: nexus-docker`
- **AND** the resolved manifest has no `credentials` key
- **THEN** `coin-lib` MUST bind Docker registry credentials using the product config value

#### Scenario: coin-lib default remains local pilot fallback

- **WHEN** product config omits an optional credential value that `coin-lib` supports through defaults
- **THEN** `coin-lib` MAY use its own defaults or Jenkins environment configuration
- **AND** `coin-api` MUST NOT provide that fallback through manifest resolve

### Requirement: Jenkins pipeline expands task DAG

coin-lib `coinPipeline` SHALL expand manifest `pipeline.tasks` into Jenkins stages using topological order of `runAfter`. Each Jenkins stage MUST invoke `coinRunStage` with task `id` matching manifest `pipeline.tasks[].id`.

#### Scenario: Linear runAfter chain

- **WHEN** manifest defines tasks `validate`, `test`, `build`, `publish` with linear `runAfter`
- **THEN** coinPipeline MUST emit Jenkins stages in DAG order
- **AND** each stage MUST call coin-executor with `--task` equal to task id

#### Scenario: Default order without runAfter

- **WHEN** manifest tasks omit `runAfter` but appear in author order
- **THEN** coinPipeline MUST treat author order as sequential dependencies for Phase A

#### Scenario: No Groovy build logic in stage expansion

- **WHEN** coinPipeline expands tasks to Jenkins stages
- **THEN** Groovy MUST NOT interpret build, containerfile, or publish logic
- **AND** MUST delegate all step execution to coin-executor

### Requirement: File-based manifest resolve in coin-lib

`coinResolveManifest` SHALL support `coin.resolve: file` per `config-resolve-file`. Remote path (API → Nexus) MUST remain the default when `resolve` is omitted or `remote`.

#### Scenario: Skip network on file resolve

- **WHEN** product config sets `coin.resolve: file` and fixture exists
- **THEN** coin-lib MUST NOT call coin-api or Nexus for that resolve
- **AND** MUST return the fixture map for materialize

#### Scenario: Remote resolve unchanged

- **WHEN** product config uses default or `coin.resolve: remote`
- **THEN** coin-lib MUST use coin-api primary and Nexus fallback as before

### Requirement: Archive test results from workspace

After a test task (or at end of pipeline when `.coin/test-results/` exists), coin-lib SHALL archive `.coin/test-results/**` as Jenkins build artifacts with `allowEmptyArchive` true. coin-lib MAY publish JUnit XML from that directory when present. Groovy MUST NOT run product tests; it only archives files produced by coin-executor.

#### Scenario: Archive test-results directory

- **WHEN** executor has written files under `.coin/test-results/`
- **THEN** coinPipeline or post-test glue MUST call Jenkins `archiveArtifacts` for `.coin/test-results/**`

#### Scenario: Empty test-results does not fail archive

- **WHEN** `.coin/test-results/` is missing or empty
- **THEN** archive step MUST NOT fail the build solely due to missing artifacts

### Requirement: Registry auth from destinations catalog

`coinConfigureRegistryAuth` SHALL configure Docker registry credentials for every host listed in schemaVersion 4 `destinations[]` where `pull` or `push` is true. Host MUST come from entry `host` when set, otherwise from `imageRegistryPrefix`. Legacy flat `destinations.imageRegistryPrefix` MUST remain supported for pre-v4 manifests.

#### Scenario: Auth covers pull and push destinations

- **WHEN** manifest destinations catalog has entry `nexus-docker` with `pull: true` and `push: true`
- **THEN** auth config MUST include that registry host before build/publish stages

#### Scenario: Legacy flat destinations still work

- **WHEN** manifest has flat `destinations.imageRegistryPrefix`
- **THEN** auth MUST configure the host derived from that prefix
