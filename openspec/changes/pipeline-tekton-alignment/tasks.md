## 1. ADR и схема

- [ ] 1.1 ADR `docs/adr/pipeline-tekton-mapping.md` — entity mapping Tekton→Coin, v4 overview
- [ ] 1.2 JSON Schema `pipeline-inline.v4.schema.json` в coin-api и `internal/gpcontent/seed/`
- [ ] 1.3 Обновить `manifest.schema.json` — `containerfiles`, `pipeline.tasks`

## 2. coin-api — validate и storage

- [ ] 2.1 Расширить `gp_release_pipeline_bodies` / validate для schemaVersion 4
- [ ] 2.2 Валидатор v4: semantic task ids, runAfter DAG, catalog refs, step kinds
- [ ] 2.3 v3→v4 migration on draft save (stages→tasks, inline containerfile→catalog)
- [ ] 2.4 v3 read adapter для resolve published releases
- [ ] 2.5 Unit-тесты validate + migration (`gpcontent`, `gp_pipeline`)

## 3. coin-api — manifest materialization

- [ ] 3.1 Manifest builder: materialize `containerfiles[]` с contentRef/digest
- [ ] 3.2 Manifest builder: `pipeline.tasks` вместо `pipeline.stages`
- [ ] 3.3 Preview API: v4 output shape + catalog в ответе
- [ ] 3.4 OpenAPI update для v4 pipeline body shape
- [ ] 3.5 Тесты builder + preview + promote

## 4. coin-executor

- [ ] 4.1 Парсинг manifest v4: `containerfiles`, `pipeline.tasks`
- [ ] 4.2 Dispatch `kind: coin` (validate, test, build, publish)
- [ ] 4.3 Dispatch `kind: containerfile` по catalog ref
- [ ] 4.4 `--task` flag (alias `--stage` deprecated)
- [ ] 4.5 Unit/integration тесты runner + manifest для v4

## 5. coin-lib

- [ ] 5.1 `coinPipeline`: topological sort `runAfter` → Jenkins stages
- [ ] 5.2 `coinRunStage`: передача `--task <id>` в executor
- [ ] 5.3 Тесты/документация glue boundary (без build logic в Groovy)

## 6. coin-ui

- [ ] 6.1 Containerfiles catalog panel на GP release detail
- [ ] 6.2 Task graph editor: tasks, runAfter, ordered steps
- [ ] 6.3 Step forms: coin / containerfile ref / sh (allowlist UI)
- [ ] 6.4 Обновить `gpContentYaml.ts` и preview для v4 shape
- [ ] 6.5 Layout: Composition → Pipeline → Containerfiles → Parameters
- [ ] 6.6 v3 draft migration UX (banner или auto-migrate on load)

## 7. Seed и контент

- [ ] 7.1 Мигрировать `coin-api/internal/gpcontent/seed/pipelines/go-app.yaml` на v4 shape
- [ ] 7.2 Мигрировать `coin-api/internal/gpcontent/seed/pipelines/go-app-docker.yaml` на v4 shape
- [ ] 7.3 Обновить `docker/scripts/seed-jenkins-lib-stack.sh` для v4
- [ ] 7.4 `make seed-jenkins-lib` — green после reseed

## 8. E2E и документация

- [ ] 8.1 E2E `demo-go-app` — validate → test → build → publish
- [ ] 8.2 E2E `demo-go-app-docker` — project containerfile catalog
- [ ] 8.3 Обновить `docs/how-to/build-stacks.md`, `docs/config.md`
- [ ] 8.4 `openspec validate pipeline-tekton-alignment --strict`
