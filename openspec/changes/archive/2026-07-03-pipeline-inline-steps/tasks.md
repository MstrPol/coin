## 1. Архитектурные решения

- [x] 1.1 Добавить ADR `docs/adr/pipeline-inline-build-stack.md`; пометить `build-stack-vnext-contract` как superseded
- [x] 1.2 Зафиксировать v3: parameters + pipeline + validateSchema; containerfile inline в buildkit steps
- [x] 1.3 Обновить `docs/how-to/build-stacks.md` под pipeline-inline (без catalog containerfiles)

## 2. coin-api: schema v3 и validation

- [x] 2.1 JSON schema v3: parameters, validateSchema, pipeline steps с inline containerfile.body / dockerfile.path
- [x] 2.2 Go model v3; отклонять `build.targets`, `deliverables`, `artifacts.containerfiles`
- [x] 2.3 Validation: buildkit steps require containerfile.body; build.id unique; publish graph
- [x] 2.4 OpenAPI preview/save под v3
- [x] 2.5 Unit tests: go-app v3, missing containerfile body, v2/vNext catalog rejection

## 3. Manifest builder и preview

- [x] 3.1 Materializer: per-step `containerfile.contentRef`+`digest`; no top-level containerfiles catalog
- [x] 3.2 manifestHash под v3 pipeline-inline shape
- [x] 3.3 Preview API для v3 structured model
- [x] 3.4 Tests: Nexus-fallback self-sufficiency, step-local containerfile refs

## 4. coin-executor runtime

- [x] 4.1 Manifest structs: inline steps с per-step containerfile ref
- [x] 4.2 Materialize containerfile from executing step (не из catalog)
- [x] 4.3 Dispatch run/build/publish inline steps
- [x] 4.4 Unit tests step-local containerfile + publish graph

## 5. coin-ui pipeline-first editor

- [x] 5.1 Секции: Parameters → Pipeline only (убрать Targets, Deliverables, Containerfiles cards)
- [x] 5.2 Buildkit step card: inline containerfile textarea + engine fields
- [x] 5.3 run/build/publish forms; publish выбирает build.id
- [x] 5.4 Client validation + preview debounce v3
- [x] 5.5 TypeScript model v3 serialize/parse
- [x] 5.6 Short hash id: генерация в UI + validate `^[a-z0-9]{5,6}$` (stage.id, build.id)

## 6. coin-gp-content и pilot reseed

- [x] 6.1 Мигрировать go-app / go-app-docker на v3 (containerfile.body в каждом buildkit step)
- [ ] 6.2 Reseed local GP `gp-01-07@1.0.0` (`cd docker && make seed-jenkins-lib` после rebuild coin-api/ui)
- [x] 6.3 Checklist: manifest preview показывает containerfile на step (gpcontent `TestPreview_inline`)

## 7. Quality gates

- [x] 7.1 Focused tests coin-api, coin-executor, coin-ui
- [x] 7.2 `openspec validate pipeline-inline-steps --strict`
- [ ] 7.3 E2E demo-go-app когда infra позволяет
