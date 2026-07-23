## 1. coin-api validate + storage

- [ ] 1.1 Validate schemaVersion 4 GP body against pipeline-inline.v4 / manifest rules
- [ ] 1.2 Persist v4 draft/release without stages/build.targets/deliverables

## 2. Materialize + preview

- [ ] 2.1 Manifest builder: emit `containerfiles[]` + `pipeline.tasks` + `destinations[]` (= fixture shape)
- [ ] 2.2 Golden test: materialize vs `samples/demo-go-app/.coin/manifest.local.yaml` shape
- [ ] 2.3 Preview API returns v4 resolved preview fields

## 3. Migration + seed

- [ ] 3.1 v3→v4 migration on draft save (stages→tasks, inline→catalog)
- [ ] 3.2 Temporary v3 read adapter for legacy releases
- [ ] 3.3 Seed `go-app` / `go-app-docker` v4; reseed PG/Nexus

## 4. coin-ui

- [ ] 4.1 Containerfiles catalog panel
- [ ] 4.2 Task graph editor (tasks, runAfter, steps, destinationRefs)
- [ ] 4.3 Layout Composition → Pipeline → Containerfiles → Parameters
- [ ] 4.4 Preview + migration UX

## 5. Remote E2E + closeout

- [ ] 5.1 E2E `demo-go-app` через `coin.resolve: remote` (без file как primary)
- [ ] 5.2 E2E `demo-go-app-docker` project containerfile
- [ ] 5.3 Sync/archive specs; закрыть change после green remote E2E
