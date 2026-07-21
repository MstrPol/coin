## 1. Specs (канон до правки docs)

- [x] 1.1 Синхронизировать main `openspec/specs/runtime-documentation` с delta (three-pin → two-pin + index)
- [x] 1.2 Добавить main `openspec/specs/docs-monorepo-layout/spec.md` из delta
- [x] 1.3 Синхронизировать `openspec/specs/gp-composition-two-slot` с delta (golden-paths consistency)

## 2. Workspace / monorepo layout

- [x] 2.1 Создать `docs/workspace-layout.md` (диаграмма sibling + `coin/` meta, таблица path→роль→corp, removed trees, seed SoT)
- [x] 2.2 Зафиксировать Q1: канон `samples/` (workspace vs `coin/samples`) в workspace-layout
- [x] 2.3 Расширить `docs/runbooks/prod-repo-split.md`: inventory sibling↔corp, без `coin-gp-content` / branching-models как extract targets; link на workspace-layout
- [x] 2.4 Обновить `coin/README.md` layout под workspace-layout

## 3. Top-level narrative (по specs)

- [x] 3.1 `docs/architecture.md` — 2-pin + embedded pipeline; убрать gp-content composition / Build Stack package narrative как live
- [x] 3.2 `docs/control-plane.md` — 2-pin, SoT layers, Platform hubs; без Studio primary
- [x] 3.3 `docs/golden-paths.md` — полный rewrite composition/authoring под `gp-release-two-pin` + `gp-embedded-pipeline`
- [x] 3.4 `docs/adr/coin-ci-runtime.md` — banner/amendment: two-pin; убрать live three-pin / gp-content pin
- [x] 3.5 `docs/adr/gp-branching-model.md` — banner composition = 2-pin (+ branching pin), не three-pin с gp-content

## 4. How-to / ops / index

- [x] 4.1 `docs/agent-build-model.md`, `jenkins-setup.md`, `responsibilities.md` — seed/make без gp-content; engine SoT = GP release
- [x] 4.2 How-to: `publish-gp-release`, `build-stacks` (deprecate/redirect), `branching-models`, onboarding — ссылки на workspace-layout
- [x] 4.3 `docs/README.md` — reading order, link workspace-layout, убрать мёртвые пути
- [x] 4.4 ADR banners где нужно: `jenkins-lib-http-nexus`, `gp-component-package-model` (live vs history)

## 5. Audit

- [x] 5.1 `rg` по `docs/` (вне явных Superseded/history): `three-pin`, `Composition (GP draft 3-pin)`, `make coin-gp-content`, `coin-gp-content/stacks`, `Component Studio` как primary path
- [x] 5.2 Проверить внутренние markdown-ссылки из `docs/README.md` и `workspace-layout.md` (нет 404 на удалённые деревья)
