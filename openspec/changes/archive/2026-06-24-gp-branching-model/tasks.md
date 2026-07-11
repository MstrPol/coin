# Tasks: gp-branching-model

**Prerequisite:** gp-component-platform ✅ (archived 2026-06-23)

## Phase 0 — ADR + composition

- [x] GBM-0.1 ADR `docs/adr/gp-branching-model.md`: 5-slot, manifest.branching, executor ownership
- [x] GBM-0.2 coin-api: GP profile + composition 5-й slot; compatibility rules
- [x] GBM-0.3 Platform lead approve 5-й slot gate

## Phase 1 — Component + API

- [x] GBM-1.1 branching-model в Component Studio (UI-03): draft → canary → stable *(scaffold GCP-1; E2E — GBM-3.4)*
- [x] GBM-1.2 coin-branching-models: trunk-based + semver-tag model.yaml + schema
- [x] GBM-1.3 coin-api manifest builder: load branching-model → manifest.branching
- [x] GBM-1.4 manifest.schema.json + OpenAPI update

## Phase 2 — Executor

- [x] GBM-2.1 coin-executor/internal/branching: ValidateBranch, ResolveVersion, ShouldPublish, Bump
- [x] GBM-2.2 Wire validate/run/version; fix --stage policy bypass
- [x] GBM-2.3 imageRef from COIN_VERSION

## Phase 3 — Pilot + tests

- [x] GBM-3.1 GP profiles: go-app* → trunk-based; второй GP → semver-tag (`DefaultBranchingModelForGP`)
- [x] GBM-3.2 seed-jenkins-lib-stack 5-slot composition
- [x] GBM-3.3 Unit tests обеих моделей (`coin-executor/internal/branching`)
- [x] GBM-3.4 E2E demo-go-app trunk-based publish policy (`docker/scripts/e2e-branching-policy.sh`)

## Phase 4 — Docs

- [x] GBM-4.1 docs/branching.md индекс каталога
- [x] GBM-4.2 docs/golden-paths.md 5-slot
- [x] GBM-4.3 coin-branching-models README для enabling team
