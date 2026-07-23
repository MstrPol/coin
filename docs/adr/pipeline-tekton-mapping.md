# ADR: Pipeline Tekton mapping (schemaVersion 4)

**Статус:** accepted (2026-07-22)  
**Change:** `pipeline-tekton-alignment`  
**Supersedes (authoring model):** [pipeline-inline-build-stack.md](./pipeline-inline-build-stack.md) (v3 stages + inline Containerfile)  
**Связанные:** [coin-ci-runtime.md](./coin-ci-runtime.md), [gp-embedded-pipeline.md](./gp-embedded-pipeline.md), [jenkins-lib-http-nexus.md](./jenkins-lib-http-nexus.md)

## Контекст

Нужна иерархия Pipeline → Task → Step, согласованная с Jenkins runtime и отладкой executor без live coin-api. Coin **не** запускает Tekton Controller.

## Решение

### Entity mapping

| Tekton | Coin (authoring + runtime) |
|--------|----------------------------|
| Pipeline | GP release pipeline / Jenkins Pipeline |
| Task | Jenkins stage (`coinRunStage` → `coin-executor --task`) |
| Step | executor action (`coin` / `containerfile` / `sh`) |
| Workspace | Jenkins workspace на dynamic agent |
| PipelineRun | Jenkins build |
| TaskRun | не моделируется |
| Triggers | вне scope |

### schemaVersion 4 (кратко)

- Top-level `containerfiles[]`: `id`, `kind: managed|project`, `path` (managed → fetch/write path).
- `pipeline.tasks[]` + `runAfter` (не `stages` для новых документов).
- Step kinds: `coin` (`validate`/`test`/`build`/`publish`), `containerfile`, `sh` (Phase A: fail-closed на execute).
- Build: `containerfileRef` = catalog id; `destinationRef` → prefix; imageRef computed; optional `cache` override; merge `.coin/outputs.json`.
- Publish: `buildTaskId` → outputs entry; `destinationRefs[]` multi-push (retag if needed).
- Destinations: top-level catalog array (auth pull|push + naming); legacy flat object only for pre-v4.
- Test: Containerfile target (default `test`); reports → `.coin/test-results/`.

### File resolve (dev)

Product config: `coin.resolve: file` + optional `coin.manifestFile` (default `.coin/manifest.local.yaml`). Soft warn; no API/Nexus. Fixture = resolved shape.

## Последствия

- Phase A: executor + coin-lib + fixture; API/UI materialize — отдельный change `pipeline-v4-control-plane`.
- Agent: single `runtime.image` (`coin-agent`); dual jnlp+buildkit pods — вне scope.
- Testcontainers / docker-socket test run — follow-up spike.
