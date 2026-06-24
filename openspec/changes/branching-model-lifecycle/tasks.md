# Tasks: branching-model-lifecycle

## 0 — Platform lead gate

- [ ] BML-0.1 Закрыть Q1/Q2 в design (scope: branching-model only vs all types; PG bodies after promote)
- [ ] BML-0.2 Amend ADR `docs/adr/gp-component-package-model.md`: canary без Nexus, Nexus на published

## 1 — coin-api: content_ref + lifecycle

- [ ] BML-1.1 `content_ref` v2: `package` optional для draft/canary; validate on promote
- [ ] BML-1.2 `register-package`: PG-only path для `branching-model` (no Nexus upload)
- [ ] BML-1.3 `promote`: Nexus upload + final content_ref для `branching-model`
- [ ] BML-1.4 Resolve materializer: canary/draft branching из PG без Nexus
- [ ] BML-1.5 OpenAPI + `schemas/content-ref.v2.schema.json` update
- [ ] BML-1.6 Unit/integration tests: register canary no Nexus, promote uploads, resolve canary from PG

## 2 — coin-ui: каталог + Studio flow

- [ ] BML-2.1 Nav + route `/branching-models` (Layout)
- [ ] BML-2.2 `BranchingModelsPage`: список моделей, версии по статусу, GP profile usage
- [ ] BML-2.3 API client: enrich endpoint или client-side join profiles + components
- [ ] BML-2.4 Component Studio: publish canary без Nexus; promote вызывает Nexus path
- [ ] BML-2.5 Ссылки catalog ↔ Studio ↔ promote wizard

## 3 — Bootstrap + docs

- [ ] BML-3.1 `publish-branching-model.sh`: draft → register (PG) → canary → promote (Nexus)
- [ ] BML-3.2 `docs/golden-paths.md` + `coin-branching-models/README.md` lifecycle update
- [ ] BML-3.3 Manual E2E script: canary resolve без Nexus blob; promote → Nexus fallback

## 4 — Acceptance

- [ ] BML-4.1 UI: каталог показывает trunk-based/semver-tag и статусы после Studio flow
- [ ] BML-4.2 API: после canary нет package в Nexus; после promote — есть
- [ ] BML-4.3 Resolve canary GP pin возвращает `manifest.branching` без coin-api недоступности на stable fallback only
