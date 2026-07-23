## Why

Нужно быстро отладить **форму resolved-манифеста и поведение coin-executor** без зависимости от coin-api / Nexus / UI. Сейчас resolve всегда ходит в сеть; pipeline-inline v3 (`stages`, inline Containerfile) плохо стыкуется с Tekton mental model (Pipeline → Task → Step). Платформенная команда должна итерировать манифест локально в репозитории.

API materialize, UI editors и remote E2E — **отдельный change** [`pipeline-v4-control-plane`](../pipeline-v4-control-plane/).

## What Changes

- **BREAKING (manifest):** `schemaVersion: 4` — `pipeline.tasks[]` (вместо `stages`), step kinds `coin` | `containerfile` | `sh`.
- **Top-level `containerfiles[]`:** peer к branching/parameters; entry = `id` + `kind: managed|project` + `path`; содержимое — файлы на диске по `path`, не inline `body` в resolved/file fixture.
- **Build/publish I/O:** imageRef computed (`destinations[]` + `destinationRef` / `destinationRefs` + `project.*` + tag); N build → merge в `.coin/outputs.json`; N publish по `buildTaskId` с multi-push.
- **Destinations catalog:** top-level `destinations[]`; build `destinationRef`; publish `destinationRefs[]`; step `cache` override; local pilot — один entry `nexus-docker`.
- **Test-in-container:** только Containerfile target; отчёты → `.coin/test-results/`; coin-lib `archiveArtifacts`.
- **config v2:** `coin.resolve: remote | file` (default `remote`); при `file` — `coin.manifestFile` (default `.coin/manifest.local.yaml`).
- **coin-lib:** file resolve → materialize `.coin/manifest.json`; soft warn; `pipeline.tasks` + `runAfter` → stages; `--task <id>`; archive test-results; registry auth по destinations catalog.
- **coin-executor:** v4 parse, catalog materialize, step dispatch, test target + report export, outputs merge, `--task`.
- **Fixture loop:** `.coin/manifest.local.yaml` для sample/demo без API/Nexus.
- ADR Tekton→Coin mapping.

**Acceptance этого change:** `demo-go-app` green на пути **file resolve** (validate → test → build, publish по branching) **без** coin-api / coin-ui / remote materialize.

## Capabilities

### New Capabilities

- `pipeline-tekton-model`: schema v4, Tekton→Coin mapping, `pipeline.tasks`, `runAfter`, step kinds.
- `gp-containerfile-catalog`: top-level `containerfiles[]` (`managed`|`project` + `path`); materialize-to-path.
- `config-resolve-file`: product config `resolve` / `manifestFile`, file-based resolve в coin-lib.

### Modified Capabilities

- `pipeline-inline-model`: supersede v3 inline-only; делегировать v4 в `pipeline-tekton-model`.
- `manifest-pipeline-inline`: resolved v4 (`tasks` + top-level catalog); file resolve path.
- `build-engine`: v4 dispatch; computed imageRef; test-via-target + `.coin/test-results/`; outputs.json; publish via `buildTaskId` + `destinationRefs`; `--task`.
- `jenkins-lib-boundary`: file resolve + task DAG → stages; archive `.coin/test-results/**`; destinations auth.

## Non-goals

- coin-api materialize / preview / migration / seed (→ `pipeline-v4-control-plane`).
- coin-ui editors (→ `pipeline-v4-control-plane`).
- Remote resolve E2E как gate этого change (→ `pipeline-v4-control-plane`).
- Замена Jenkins на Tekton Controller / CRD YAML / Triggers.
- Hard guardrail на `resolve: file` (только soft warn).
- Corp fleet migration prod GP.
- Jenkins `parallel` по `runAfter` frontier.
- Testcontainers / test `podman run` + docker socket.

## Impact

| Область | В scope |
|---------|---------|
| **config v2 schema** | `resolve`, `manifestFile` |
| **coin-lib** | file resolve, `--task`, task stages, destinations auth |
| **coin-executor** | v4 parse + dispatch |
| **samples / docs** | fixture + how-to local manifest; offline Jenkins acceptance |
| **coin-api** | только schema artifact (контракт), без storage/builder |
| **coin-ui** | out of scope |
| **docs/adr** | pipeline-tekton-mapping |
