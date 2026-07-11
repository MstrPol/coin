# coin-executor-agent-pipeline Specification

## Purpose
TBD - created by archiving change coin-executor-agent-image-job. Update Purpose after archive.
## Requirements
### Requirement: coin-executor job builds and publishes coin-agent image

Jenkins job `coin-executor` SHALL build `coin-agent` container image from `Dockerfile.agent` using multi-stage Docker build and SHALL push the resulting image to Nexus Docker repository.

The image build SHALL compile `coin-executor` binary inside a golang builder stage and bake it into the runtime stage based on `jenkins/inbound-agent` with `podman` and BuildKit.

#### Scenario: Successful multi-stage image publish
- **WHEN** publisher starts `coin-executor` job with valid parameters and credentials
- **THEN** pipeline MUST run `docker build -f Dockerfile.agent` for selected `GOARCH`
- **AND** MUST produce `coin-agent:<version>` with baked `coin-executor` binary
- **AND** MUST push the image to `${NEXUS_DOCKER_REPO}` in Nexus Docker registry

#### Scenario: Fail if docker runtime unavailable
- **WHEN** job starts on Jenkins node without working Docker daemon access
- **THEN** pipeline MUST fail before publish stage
- **AND** MUST log a diagnostic message describing missing Docker access

### Requirement: coin-executor job registers agent draft in platform

After successful image push, `coin-executor` job SHALL register `agent/coin-agent@<version>` as `draft` through coin-api Admin API using agent metadata from the published image.

#### Scenario: Register draft after push
- **WHEN** image push succeeds for version `<version>`
- **THEN** pipeline MUST call coin-api draft create endpoint for `agent/coin-agent`
- **AND** metadata MUST include image reference and digest when available
- **AND** resulting component version MUST stay in `draft` status

#### Scenario: Publish stage fails on API register error
- **WHEN** coin-api returns non-success response on draft create request
- **THEN** pipeline MUST mark build as failed
- **AND** MUST include HTTP code and response body in logs

### Requirement: Jenkinsfile remains orchestration-only for agent publish

`coin-executor/Jenkinsfile` SHALL orchestrate versioning, credential binding, optional `go test`, and stage sequencing, while image build/publish business logic SHALL execute via repository script `scripts/publish-agent.sh`.

Jenkinsfile MUST NOT compile production binary outside `Dockerfile.agent` multi-stage build.

#### Scenario: Pipeline delegates image build to Dockerfile.agent
- **WHEN** publish stage runs
- **THEN** Jenkinsfile MUST invoke `scripts/publish-agent.sh` with resolved version and architecture
- **AND** `publish-agent.sh` MUST call `docker build -f Dockerfile.agent` on repository root
- **AND** MUST NOT inline docker build/push and API payload logic in Groovy

#### Scenario: Docker login uses writable config directory
- **WHEN** publish stage logs in to Nexus Docker registry from Jenkins worker
- **THEN** script MUST use writable `DOCKER_CONFIG` (for example `${WORKSPACE}/.docker`)
- **AND** MUST NOT require write access to `/root/.docker`

