## 1. Архитектурные решения и контракт

- [x] 1.1 Закрыть BLOCKING вопросы из `design.md`: YAML export/debug, parameter types, manifest versioning, несколько deliverables одного type
- [x] 1.2 Добавить ADR `docs/adr/build-stack-vnext-contract.md`
- [x] 1.3 Обновить документацию hard GP: Build Stack = параметры, targets, deliverables, Containerfiles, pipeline contract
- [x] 1.4 Зафиксировать migration note: `gp-manifest-publish-routing` архивирован без sync как промежуточный change

## 2. coin-api model, schema и package validation

- [x] 2.1 Добавить Go model для Build Stack vNext в gp-content package
- [x] 2.2 Добавить JSON schema для Build Stack vNext
- [x] 2.3 Реализовать validation ссылок: `targetId`, `deliverableId`, parameter refs, Containerfile refs
- [x] 2.4 Запретить credentials/secret values в parameters
- [x] 2.5 Добавить поддержку named Containerfile artifacts в gp-content package/content refs
- [x] 2.6 Обновить OpenAPI для Build Stack vNext draft/update/preview responses
- [x] 2.7 Добавить unit tests для valid/invalid Build Stack vNext packages

## 3. Manifest builder и preview

- [x] 3.1 Материализовать `parameters`, `build.targets`, `deliverables`, `artifacts.containerfiles`, `pipeline.stages` в resolved manifest
- [x] 3.2 Убрать top-level `build.engine` как dispatch source для Build Stack vNext
- [x] 3.3 Обновить `manifestHash`, чтобы он учитывал targets/deliverables/parameters/artifacts/stages
- [x] 3.4 Переделать gp-content preview API под canonical model и единый resolved manifest preview
- [x] 3.5 Добавить schema/tests: manifest содержит vNext sections и не содержит credentials

## 4. coin-ui Build Stack vNext editor

- [x] 4.1 Переделать create/edit экран Build Stack на model-first sections
- [x] 4.2 Добавить Parameters card: type/default/required/allowed values
- [x] 4.3 Добавить Targets card: dynamic targets и engine-specific fields per target
- [x] 4.4 Добавить Deliverables card: dynamic deliverables, type, target, publish settings
- [x] 4.5 Добавить Containerfile artifacts card: named managed artifacts и привязка к targets
- [x] 4.6 Переделать Pipeline stages card: typed stages с platform action steps
- [x] 4.7 Убрать дублирующий YAML/JSON primary preview; оставить resolved manifest preview и validation issues
- [x] 4.8 Добавить client-side validation и UI tests/build check

## 5. coin-executor runtime

- [x] 5.1 Обновить manifest structs под Build Stack vNext
- [x] 5.2 Реализовать parameter resolution без secrets/credentials
- [x] 5.3 Реализовать target-level engine dispatch для `buildkit` и `dockerfile`
- [x] 5.4 Реализовать materialization named Containerfile artifacts
- [x] 5.5 Реализовать execution stage steps: `run-target`, `build-deliverable`, `publish-deliverable`
- [x] 5.6 Обновить publish refs: использовать destinations + project identity + deliverable metadata
- [x] 5.7 Добавить unit tests для targets, parameters, deliverables и stage steps

## 6. Seed data, samples и local pilot

> **Статус change:** implementation P0 завершён; pilot/E2E задачи 6.2–6.3, 6.6, 6.8 **отложены и superseded** change `pipeline-inline-steps` (вариант C: inline config в pipeline steps). Change архивируется без sync delta specs.

- [x] 6.1 Обновить `coin-gp-content` pilot stacks на Build Stack vNext
- [ ] 6.2 Пересеять local GP `gp-01-07@1.0.0` с Build Stack vNext *(superseded: pipeline-inline-steps)*
- [ ] 6.3 Проверить resolved manifest: vNext sections есть, top-level credentials нет, `coin-lib` не хранит build logic *(superseded)*
- [x] 6.4 Обновить `samples/demo-go-app` только при необходимости, сохраняя thin product config
- [x] 6.5 Запустить focused tests для `coin-api`, `coin-executor`, `coin-ui`
- [ ] 6.6 Запустить E2E `samples/demo-go-app`: resolve → validate → test → build → publish *(superseded)*
- [ ] 6.8 Ручная проверка Build Stack vNext editor в coin-ui (`/platform/build-stacks/:name/:version/edit`) *(superseded — UI будет переделан)*
- [x] 6.7 Запустить `openspec validate rework-build-stack-contract --strict`
