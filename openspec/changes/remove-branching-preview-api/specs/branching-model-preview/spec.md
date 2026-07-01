## REMOVED Requirements

### Requirement: Branching model preview API

**Reason**: The preview endpoint forced `coin-api` to import `coin-executor` branching logic. Branching behavior evaluation belongs to `coin-executor` runtime, while `coin-api` should store and resolve branching model metadata without depending on executor.

**Migration**: Remove `POST /v1/admin/branching-models/preview` from API, OpenAPI, and UI clients. Operators continue to author branching models through schema/card/YAML editing and validate/register/promote lifecycle. Runtime branching behavior is verified by `coin-executor` tests and CI execution.

coin-api SHALL expose `POST /v1/admin/branching-models/preview` to evaluate branching policy scenarios using coin-executor branching logic.

The endpoint MUST NOT be merged into validate-package.

#### Scenario: Preview branch match and version

- **WHEN** publisher sends a schema v2 model and scenario `{ branch: "feature/PROJ-101" }`
- **THEN** the response MUST include matched rule name, `coinVersion`, and template evaluation result or error

#### Scenario: Preview publish eligibility with request

- **WHEN** scenario includes `requestPublish: true` and branch matches a rule with `publish: false`
- **THEN** the response MUST indicate publish denied (not allowed)

#### Scenario: Preview publish allowed

- **WHEN** scenario includes `requestPublish: true` and branch matches a rule with `publish: true`
- **THEN** the response MUST indicate publish allowed

#### Scenario: Executor is source of truth

- **WHEN** preview is evaluated
- **THEN** coin-api MUST use coin-executor branching package logic (not client-side reimplementation)
