## Phase A — executor + file resolve (весь AC этого change)

### 1. ADR и схемы

- [x] 1.1 ADR `docs/adr/pipeline-tekton-mapping.md` — entity mapping Tekton→Coin, v4 overview, file-resolve note
- [x] 1.2 JSON Schema `pipeline-inline.v4.schema.json` (coin-api tree и/или `docs/schemas` / seed — артефакт контракта)
- [x] 1.3 Обновить `manifest.schema.json` — `containerfiles`, `pipeline.tasks` (вместо/наряду со `stages` для v4)
- [x] 1.4 Обновить `config.v2.schema.json` — `coin.resolve` (`remote`|`file`), `coin.manifestFile`

### 2. config + coin-lib resolve

- [x] 2.1 `coinResolveManifest`: ветка `resolve: file` → читать `manifestFile` (default `.coin/manifest.local.yaml`)
- [x] 2.2 Soft warn в логе при file resolve; не ходить в API/Nexus
- [x] 2.3 Materialize `.coin/manifest.json` после file resolve (существующий `coinMaterializeDotCoin`)
- [x] 2.4 Docs: `docs/config.md` — `resolve` / `manifestFile`; how-to local manifest

### 3. coin-executor

- [x] 3.1 Парсинг manifest v4: top-level `containerfiles` (`id`/`kind`/`path`), `pipeline.tasks`
- [x] 3.2 Materialize catalog: `managed` → fetch/write `path`; `project` → use `path`
- [x] 3.3 Dispatch `kind: coin` validate; test via Containerfile target (default `test`); build по `containerfileRef` + computed imageRef
- [x] 3.4 Test: export отчётов в `.coin/test-results/` (export stage / `--output`); best-effort при fail
- [x] 3.5 Build merge entry в `.coin/outputs.json` (`name` = task id); N builds → один файл
- [x] 3.6 Dispatch publish по `buildTaskId` + `destinationRefs[]` → multi-push; branching/publish gates
- [x] 3.7 Dispatch `kind: containerfile` / `kind: sh` (sh не для default test; allowlist или fail closed)
- [x] 3.8 `--task` flag; `--stage` deprecated alias
- [x] 3.9 Unit/integration тесты: catalog, test target + reports dir, multi-build outputs, publish by id
- [x] 3.10 Destinations catalog: parse/validate array; build `destinationRef`; publish `destinationRefs`; `cache` override

### 4. coin-lib pipeline

- [x] 4.1 `coinPipeline`: читать `pipeline.tasks` (compat: v3 `stages` если ещё нужен remote pilot)
- [x] 4.2 Topological / author-order → Jenkins stages
- [x] 4.3 `coinRunStage`: передача `--task <id>` в executor
- [x] 4.4 Archive `.coin/test-results/**` (`allowEmptyArchive: true`); опционально junit
- [x] 4.5 Документация glue boundary (без build logic / без test cmd в Groovy)
- [x] 4.6 `coinConfigureRegistryAuth`: login всем destinations с `pull|push` (не только flat prefix)

### 5. Fixture и acceptance (без API / UI)

- [x] 5.1 Канонический `.coin/manifest.local.yaml` + Containerfile с `test` / `test-reports` stages
- [x] 5.2 Sample config: `coin.resolve: file` + `goldenPath`/`version`
- [x] 5.3 Fixture destinations: один entry `nexus-docker` (pull+push); build `destinationRef`; publish `destinationRefs`
- [ ] 5.4 Acceptance: `demo-go-app` validate → test → build (+ publish по branching) через **file resolve**, без coin-api / coin-ui / remote materialize (offline validate ✅; полный Jenkins — после `publish-agent` + push sample)
- [x] 5.5 `openspec validate pipeline-tekton-alignment --strict`

> API / UI / remote E2E — change [`pipeline-v4-control-plane`](../pipeline-v4-control-plane/).
