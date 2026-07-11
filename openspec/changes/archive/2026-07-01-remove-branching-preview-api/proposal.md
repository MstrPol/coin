## Why

`coin-api` currently depends on `coin-executor` only to power the branching-model preview endpoint. This reverses the intended boundary: `coin-api` should store and resolve platform metadata, while `coin-executor` is the runtime authority that evaluates branching rules during CI.

The preview feature is not required for the local pilot E2E path. Removing it is the smallest way to restore module independence without adding a new shared package or duplicating runtime branching logic in `coin-api`.

## What Changes

- **BREAKING**: Remove `POST /v1/admin/branching-models/preview` from the Admin API and OpenAPI contract.
- Remove `coin-api` dependency on `coin.local/coin-executor` and the local `replace` directive.
- Remove branching-model preview UI calls and result panels from the Platform branching model editor.
- Keep branching model authoring, YAML/card editing, validate/register/promote, resolve materialization, and executor runtime enforcement unchanged.
- Update docs/specs to state that `coin-executor` is the only branching behavior evaluator on the build path.

## Non-goals

- Do not introduce a new shared Go module/package for branching logic.
- Do not copy executor branching runtime logic into `coin-api`.
- Do not remove GP resolve preview or gp-content preview.
- Do not change branching model schema v2, lifecycle, storage, or manifest materialization.
- Do not change executor versioning or publish policy behavior.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `branching-model-preview`: remove the Admin preview API capability.
- `branching-models-catalog`: remove UI requirements that depend on branching preview API feedback.

## Impact

- `coin-api`: remove preview route/handler/tests/OpenAPI schemas and the `coin-executor` module dependency.
- `coin-ui`: remove `branchingModelPreview` client usage and preview/test UI state from branching model editor.
- `coin-executor`: no runtime behavior change; branching package remains executor-owned.
- Docs: update branching model authoring docs and user guide to remove platform preview references.
