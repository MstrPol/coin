## 1. coin-api Dependency Boundary

- [x] 1.1 Remove `POST /v1/admin/branching-models/preview` route and handler from `coin-api`
- [x] 1.2 Remove branching preview API tests and OpenAPI path/schemas
- [x] 1.3 Remove `coin.local/coin-executor` require/replace from `coin-api/go.mod` via `go mod tidy`
- [x] 1.4 Verify no `coin-api` package imports `coin.local/coin-executor`

## 2. coin-ui Branching Editor

- [x] 2.1 Remove `branchingModelPreview` API client method and related types
- [x] 2.2 Remove branch test/scenario preview state, controls, and result panel from `BranchingModelEditor`
- [x] 2.3 Keep rule cards, YAML mapping, validation/register/promote, and docs links working

## 3. Documentation and Specs

- [x] 3.1 Update branching model docs/user guide to remove preview API references
- [x] 3.2 Update any design/runbook text that says branching preview is executor-backed
- [x] 3.3 Validate OpenSpec change and ensure delta specs reflect removed preview capability

## 4. Verification

- [x] 4.1 Run relevant `coin-api` tests and confirm module builds without executor dependency
- [x] 4.2 Run relevant `coin-ui` typecheck/tests or focused build checks
- [x] 4.3 Search repo for remaining active references to `/v1/admin/branching-models/preview`
- [x] 4.4 Confirm local pilot build path remains unchanged: resolve emits `manifest.branching`, executor owns runtime evaluation
