## Context

Branching model schema v2 added an Admin preview endpoint so the Platform editor could evaluate example branch scenarios. That endpoint imports `coin.local/coin-executor/pkg/branching`, which forces `coin-api` to depend on the executor module.

This conflicts with the current runtime boundary:

```
coin-api      -> stores components, resolves manifest
coin-executor -> evaluates manifest.branching during CI
coin-lib      -> Jenkins glue only
```

Preview is operator UX, not build-path behavior. Local pilot acceptance depends on `coin-executor` enforcing `manifest.branching`, not on `coin-api` simulating branching decisions.

## Goals / Non-Goals

**Goals:**

- Remove `coin-api -> coin-executor` Go module dependency.
- Remove branching-model preview API and UI features that depend on it.
- Keep branching model authoring as schema/card/YAML editing plus validate/register/promote.
- Keep executor branching runtime behavior unchanged.

**Non-Goals:**

- No new shared branching package/module.
- No duplicated branching evaluator inside `coin-api`.
- No changes to GP resolve preview or gp-content preview.
- No changes to branching model schema v2 or manifest `branching` materialization.

## Decisions

### D1. Remove platform branching preview instead of sharing evaluator code

We choose to remove `POST /v1/admin/branching-models/preview` and its UI consumers.

Alternatives considered:

| Option | Result | Decision |
|--------|--------|----------|
| Keep `coin-api -> coin-executor` | Preserves preview but violates module boundary | Rejected |
| Add shared branching package/module | Clean logic sharing but increases package/module count | Rejected for local pilot |
| Copy evaluator into `coin-api` | Removes dependency but risks drift from runtime | Rejected |
| Remove preview | Restores boundary with minimal moving parts | Accepted |

### D2. Validation remains structural, execution remains runtime-only

`coin-api` should validate draft packages structurally and store artifact bodies. It should not evaluate branch/version/publish behavior. Runtime behavior remains in `coin-executor`, which consumes the resolved manifest during CI.

### D3. UI keeps authoring help, not executable preview

The branching model editor can keep:

- ordered rule cards;
- YAML preview/edit mapping;
- docs links;
- validate/register/promote lifecycle actions.

It must remove scenario preview, test branch calls, and any UI text that implies platform preview is executor-backed.

## Risks / Trade-offs

- Reduced operator UX while editing branching rules -> mitigate with clearer docs and validation errors.
- Mistakes may be discovered later during CI -> keep executor branching tests and sample E2E coverage.
- Archived specs/docs may still mention preview -> update active specs and current docs; archived changes remain historical.

## Migration Plan

1. Remove preview API route, handler, tests, OpenAPI path/schemas, and `coin-executor` dependency from `coin-api/go.mod`.
2. Remove branching preview client method and UI panel from `coin-ui`.
3. Update branching model docs/user guide.
4. Run `go mod tidy` for `coin-api`.
5. Verify `coin-api` builds/tests without `coin.local/coin-executor` in `go.mod`.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Нужно ли сохранять preview через shared package? | ✅ | A: shared package; B: remove preview | B: remove preview, чтобы минимизировать пакеты и убрать зависимость |
