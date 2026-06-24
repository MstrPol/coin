# Tasks: branching-model-lifecycle

## 0 — Platform lead gate

- [x] BML-0.1 Закрыть Q1/Q2 в design (scope: branching-model only vs all types; PG bodies after promote)
- [x] BML-0.2 Amend ADR `docs/adr/gp-component-package-model.md`: canary без Nexus, Nexus на published

## 1 — coin-api: content_ref + lifecycle

- [x] BML-1.1 `content_ref` v2: `package` optional для draft/canary; validate on promote
- [x] BML-1.2 `register-package`: PG-only path для `branching-model` (no Nexus upload)
- [x] BML-1.3 `promote`: Nexus upload + final content_ref для `branching-model`
- [x] BML-1.4 Resolve materializer: canary/draft branching из PG без Nexus
- [x] BML-1.5 OpenAPI + `schemas/content-ref.v2.schema.json` update
- [x] BML-1.6 Unit/integration tests: register canary no Nexus, promote uploads, resolve canary from PG

## 2 — coin-ui: каталог + Studio flow

- [x] BML-2.1 Nav + route `/branching-models` (Layout)
- [x] BML-2.2 `BranchingModelsPage`: список моделей, версии по статусу, GP profile usage
- [x] BML-2.3 API client: enrich endpoint или client-side join profiles + components
- [x] BML-2.4 Component Studio: publish canary без Nexus; promote вызывает Nexus path
- [x] BML-2.5 Ссылки catalog ↔ Studio ↔ promote wizard

## 3 — Bootstrap + docs

- [x] BML-3.1 `publish-branching-model.sh`: draft → register (PG) → canary → promote (Nexus)
- [x] BML-3.2 `docs/golden-paths.md` + `coin-branching-models/README.md` lifecycle update
- [x] BML-3.3 Manual E2E script: canary resolve без Nexus blob; promote → Nexus fallback

## 4 — Acceptance

- [x] BML-4.1 UI: каталог показывает trunk-based/semver-tag и статусы (`/branching-models`, API verified)
- [x] BML-4.2 API: после canary нет package в Nexus; после promote — есть (`e2e-branching-model-lifecycle.sh`)
- [x] BML-4.3 Resolve canary GP pin возвращает `manifest.branching` из PG (`e2e-branching-canary-resolve.sh`)
