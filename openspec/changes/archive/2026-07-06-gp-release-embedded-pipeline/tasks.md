## 1. Архитектурные решения

- [x] 1.1 ADR `docs/adr/gp-embedded-pipeline.md` (accepted); supersede gp-content section в `gp-component-package-model.md`
- [x] 1.2 Зафиксировать hard cut: нет `gp-content` registry, 2-pin composition, pipeline на GP release
- [x] 1.3 Обновить `docs/golden-paths.md`, `docs/how-to/publish-gp-release.md`, user guide под GP hub authoring

## 2. coin-api: storage и migration

- [x] 2.1 Migration: pipeline body table (или `gp_artifact_bodies` key `pipeline`); purge `gp-content` из `component_versions` и `gp_composition`
- [x] 2.2 Store/load embedded pipeline-inline v3 на GP release draft
- [x] 2.3 Убрать `gpContentName` из draft/create/patch API; composition только `agent` + `branching-model`
- [x] 2.4 Promote gate: valid pipeline + published external pins (без gp-content check)
- [x] 2.5 Unit tests: 2-pin draft, reject gp-content composition, pipeline validation

## 3. coin-api: preview, resolve, OpenAPI

- [x] 3.1 `POST .../golden-paths/{name}/versions/{version}/pipeline/preview` (перенос логики gp-content preview)
- [x] 3.2 Resolve: materialize pipeline из GP release body + 2 pins; убрать gp-content loader path
- [x] 3.3 Manifest builder: без gp-content package lookup для published release
- [x] 3.4 OpenAPI: новые GP pipeline endpoints; удалить gp-content admin paths
- [x] 3.5 Tests: resolve go-app / go-app-docker с embedded pipeline, manifestHash на containerfile change

## 4. coin-api: cleanup

- [x] 4.1 Удалить component handlers/validate для type `gp-content`
- [x] 4.2 Обновить seed (`gpcontent/seed*.go`): pipeline в GP releases, не component registry
- [x] 4.3 Убрать alias profiles `xxx` / `gp-01-07` из seed scripts

## 5. coin-ui: GP release pipeline editor

- [x] 5.1 Перенести pipeline editor на GP release detail (`GpReleasePipelineEditor`)
- [x] 5.2 Preview debounce через GP release pipeline preview API
- [x] 5.3 Draft wizard и composition form: 2 picker'а (agent, branching); scaffold pipeline на new draft
- [x] 5.4 Promote UX: блокировка только по external draft pins + pipeline errors
- [x] 5.5 Удалить `/platform/build-stacks/*`, family config, nav entries

## 6. coin-gp-content и bootstrap

- [x] 6.1 `coin-gp-content/stacks/*` — seed source only; deprecate `publish-content.sh` как primary path
- [x] 6.2 `docker/scripts/seed-jenkins-lib-stack.sh`: GP profiles `go-app`, `go-app-docker` с embedded pipeline
- [ ] 6.3 Reseed local pilot (`make seed-jenkins-lib`)

## 7. Quality gates

- [x] 7.1 Focused tests coin-api, coin-ui
- [x] 7.2 `openspec validate gp-release-embedded-pipeline --strict`
- [ ] 7.3 E2E `demo-go-app` и `demo-go-app-docker` green
