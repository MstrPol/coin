## 1. coin-api — profile metadata

- [x] 1.1 Migration: `gp_profiles` — add `description`, drop `slots`; audit/backfill strategy for pilot
- [x] 1.2 `CreateGPProfile` / `GET profile` — `name` + optional `description` only; OpenAPI update
- [x] 1.3 Remove `ValidateCanonicalGPSlots` from profile create path; update admin handlers
- [x] 1.4 Unit tests for profile CRUD without slots

## 2. coin-api — composition & platform runtime

- [x] 2.1 Extend `platform_settings` (or runtime endpoint) with agent/executor/lib pins
- [x] 2.2 Draft/release validation: accept only `gp-content` + `branching-model`; reject platform keys in new drafts
- [x] 2.3 Resolve: merge platform runtime line + GP composition → full manifest
- [x] 2.4 Integration tests: two-slot draft → promote → resolve green
- [x] 2.5 Bootstrap/seed: platform runtime defaults + metadata-only GP profiles

## 3. coin-ui — profile & catalog

- [x] 3.1 `CreateGPProfile`: form name + description only; remove slot/version pickers
- [x] 3.2 `GpCatalogPage`: remove Slots column; add description column (truncated)
- [x] 3.3 API types + client for new profile shape

## 4. coin-ui — draft-only publish flow

- [x] 4.1 Remove `/gp/:name/releases/new`, hub «New release», GpNewRelease wrapper
- [x] 4.2 Draft wizard: composition picker for `gp-content` + `branching-model` only
- [x] 4.3 Fix PublishWizard scoped mode — no «Нет Golden Path» when `scopedGpName` set
- [x] 4.4 Legacy redirects: `/releases/publish` → `new-draft`; `/releases/new` route removed
- [x] 4.5 `GpOverviewTab`: remove profile slots table; welcome → first draft CTA
- [x] 4.6 `GpHubLayout` / Releases tab: single «New draft» action

## 5. Platform runtime UI (minimal)

- [x] 5.1 Platform settings or Runtime page: view/edit agent, executor, lib pins (admin)
- [x] 5.2 Read-only display of current platform line for readers

## 6. Docs, E2E, validate

- [x] 6.1 Update `docs/golden-paths.md` composition section (2-slot GP + platform runtime)
- [x] 6.2 Update `coin-ui/README.md`, `docs/coin-ui-user-guide.md`
- [ ] 6.3 E2E: `demo-go-app` green after wipe-gp + bootstrap
- [x] 6.4 `openspec validate gp-profile-metadata-model --strict`

## 7. Composition catalog decoupling (follow-up — spec D2/D8)

> **Контекст:** pilot реализовал `gp-content` name = profile name; product model — catalog pickers. Требует platform lead sign-off на `gpContentName` API (Q5).

- [x] 7.1 coin-api: `gpContentName` в draft/publish body; `validateNewGPComposition(gpContentName, branchingModelName, …)`; OpenAPI
- [x] 7.2 coin-api: `gp_composition` хранит фактические component names; resolve без привязки к profile name
- [x] 7.3 coin-ui: draft wizard — dropdown gp-content **name** + version (все stacks из registry)
- [x] 7.4 coin-ui: hub Build stack tab — gp-content из latest composition *(superseded §10 — tab удаляется)*
- [x] 7.5 docs: `golden-paths.md` — lifecycle platform components → GP draft
- [x] 7.6 tests + seed: profile `xxx` + `gpContentName: go-app` E2E path

## 8. Draft deletion (spec D9)

- [x] 8.1 coin-api: `DELETE /golden-paths/{name}/versions/{version}` — только `status=draft`; 409 для published; audit `delete_gp_draft`
- [x] 8.2 coin-api: cascade delete composition + draft artifact bodies; OpenAPI
- [x] 8.3 coin-ui: «Delete draft» на release detail + Releases tab (publisher+); confirm dialog
- [x] 8.4 docs: golden-paths / user guide — draft vs published immutability

## 9. Three-pin draft composition (spec D2/D3/D8 — supersedes §7 two-slot)

> **Контекст:** §7 реализовал 2-slot + platform agent/executor/lib. Product: draft = agent stack + branching-model + gp-content; platform = lib only.

- [x] 9.1 coin-api: `agentStackName` + `agent` key в composition; validation 3-pin; reject `lib`/`executor` keys; resolve executor from agent
- [x] 9.2 coin-api: `platform_settings.runtime` — lib only; migration shrink agent/executor defaults
- [x] 9.3 coin-ui: draft wizard — 3 pickers (agent, branching-model, gp-content); platform settings lib-only
- [x] 9.4 docs + seed: 3-pin body; golden-paths lifecycle update
- [x] 9.5 tests: draft with agent+gp-content+branching → resolve manifest green

## 10. Remove profile Build stack tab (spec D10)

> **Контекст:** gp-content pin живёт в composition версии GP, не в профиле. Hub tab создаёт ложную связь profile ↔ build stack.

- [x] 10.1 coin-ui: убрать вкладку Build stack, route `/gp/:name/build-stack`, `GpBuildStackTab`
- [x] 10.2 coin-ui: release detail — composition table с deep link Studio/Platform для gp-content; убрать ссылку «Build stack →» на hub tab
- [x] 10.3 coin-ui: Overview / GpReleaseDetail — убрать ссылки на `/build-stack`
- [x] 10.4 docs: user guide, golden-paths — gp-content на release detail; Platform catalog для stacks
- [x] 10.5 `openspec validate gp-profile-metadata-model --strict`

## 11. Draft wizard layout + edit composition (spec D11/D12)

> **Контекст:** форма new draft — съехавшие колонки; draft должен редактироваться до promote, published — read-only.

- [x] 11.1 coin-ui: `GpCompositionForm` — таблица Slot / Component / Version; использовать в new-draft wizard
- [x] 11.2 coin-api: `PATCH /golden-paths/{name}/versions/{version}` — только draft; 409 published; audit `update_gp_draft`
- [x] 11.3 coin-ui: release detail draft — editable composition + Save; published read-only
- [x] 11.4 docs: user guide — draft edit vs published immutability
- [x] 11.5 `openspec validate gp-profile-metadata-model --strict`
