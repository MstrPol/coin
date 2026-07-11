## 1. Schema v2 и каталог

- [x] 1.1 `branching-model.schema.json` v2 (branches, template, publish)
- [x] 1.2 Эталоны `coin-branching-models/models/trunk-based/model.yaml` и `semver-tag/model.yaml` на v2
- [x] 1.3 Сжатые README per model + scenarios table для UI presets
- [x] 1.4 Удалить/заменить v1 schema references

## 2. coin-executor — branching engine v2

- [x] 2.1 MatchRule: ordered branches, first match, RE2 named captures
- [x] 2.2 ResolveVersion from `versioning.template` + `{base}`, `{jira}`, `{n}`, `{branch}`
- [x] 2.3 `AllowsPublish(rule, COIN_PUBLISH_REQUEST)` — fail when requested + denied
- [x] 2.4 Удалить v1 paths (`branchTypes`, `publish.when`, `isSemverTagModel` by name)
- [x] 2.5 `PreviewScenarios` для coin-api
- [x] 2.6 Unit tests: trunk-based presets, publish fail on feature+request, main/master rule

## 3. coin-api

- [x] 3.1 validate-package: только schema v2
- [x] 3.2 `POST /v1/admin/branching-models/preview` + OpenAPI schemas
- [x] 3.3 manifest.branching: pass-through `branches[]`
- [x] 3.4 Server test smoke

## 4. coin-lib

- [x] 4.1 При `params.publish=true` выставлять `COIN_PUBLISH_REQUEST=true` в pod env
- [x] 4.2 PR branch normalization: prefer `CHANGE_BRANCH` when set

## 5. coin-ui — rule builder + preview

- [x] 5.1 Заменить `BranchingModelEditor` на v2 branch cards + template field
- [x] 5.2 Двухколоночный layout + YAML preview
- [x] 5.3 API client `branchingModelPreview()`
- [x] 5.4 Test branch name + scenario panel (`requestPublish`)
- [x] 5.5 Удалить v1 types/parser в `branchingModelYaml.ts`

## 6. Docs

- [x] 6.1 Удалить `docs/branching.md`, починить все ссылки
- [x] 6.2 Создать `docs/how-to/branching-models.md`
- [x] 6.3 Обновить `docs/README.md`, `coin-ui-user-guide.md`, `docs/openapi.md`
- [x] 6.4 `docs/how-to/add-new-service-repo.md` — pointer на GP branching model

## 7. Docker / E2E

- [x] 7.1 Обновить `seed-jenkins-lib-stack.sh` / publish scripts под v2
- [x] 7.2 Обновить `e2e-branching-policy.sh` под publish request + fail semantics
- [x] 7.3 Удалить orphan dirs `openspec/changes/branching-docs-catalog`, `branching-editor-preview` если не нужны

## 8. Validation

- [x] 8.1 `go test ./...` coin-api, coin-executor
- [x] 8.2 `npm run build` coin-ui
- [x] 8.3 `openspec validate branching-model-schema-v2 --strict`
- [x] 8.4 `make e2e-demo-go-app` или `e2e-branching-policy` green
