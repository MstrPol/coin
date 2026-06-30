## 1. Schema contract

- [x] 1.1 `schemas/gp-content.schema.json` v2 (два engine, artifacts, capabilities, без controls/when)
- [x] 1.2 Эталон `coin-gp-content/stacks/go-app/content.yaml` → v2
- [x] 1.3 Новый stack `go-app-docker` (BYO dockerfile); удалить `go-app-bp`, `go-app-df`

## 2. coin-api validate + preview

- [x] 2.1 `validate_draft.go` — schema v2 rules, engine-specific artifacts/deliverables
- [x] 2.2 Reject `buildpack`, v1 content without schemaVersion
- [x] 2.3 `POST /v1/admin/gp-content/preview` + OpenAPI
- [x] 2.4 `manifest.Builder` — BYO dockerfile path (no containerfile ref)

## 3. coin-executor + agent

- [x] 3.1 Удалить buildpack dispatch и tests
- [x] 3.2 BYO dockerfile: build/test/publish из workspace path
- [x] 3.3 buildkit: materialize Containerfile без изменений semantics
- [x] 3.4 `Dockerfile.agent`: убрать pack, paketo-builder.tar; coinPipeline bootstrap без buildpack branch

## 4. coin-ui editor

- [x] 4.1 Rewrite `gpContentYaml.ts` (bijective v2, два engine)
- [x] 4.2 `GpContentEditor` cards + preview panel (как BranchingModelEditor)
- [x] 4.3 Presets go-app / go-app-docker; capabilities enforce
- [x] 4.4 `buildGpContentManifestSubset` — capabilities + v2 shape

## 5. Seed, samples, E2E

- [x] 5.1 `seed-jenkins-lib-stack.sh` — без go-app-bp; go-app-docker sample GP
- [x] 5.2 Sample repo `demo-go-app-docker` + Jenkins job
- [x] 5.3 `make e2e-build-engines` — 2 jobs (buildkit + BYO)
- [x] 5.4 Удалить demo-go-app-bp / demo-go-app-df references

## 6. Docs + ADR

- [x] 6.1 Amend `build-engine-contract.md`, `coin-ci-runtime.md` (2 engines, no buildpack)
- [x] 6.2 `docs/how-to/build-stacks.md` (новый); update agent-build-model, golden-paths
- [x] 6.3 `openspec validate gp-content-schema-v2 --strict`

## 7. Archive

- [x] 7.1 Archive change; baseline specs; Purpose updates
