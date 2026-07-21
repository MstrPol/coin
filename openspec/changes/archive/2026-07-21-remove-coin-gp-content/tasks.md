## 1. Safety gate (seed parity)

- [x] 1.1 Diff `coin-gp-content/stacks/go-app/content.yaml` ↔ `coin-api/internal/gpcontent/seed/pipelines/go-app.yaml` (и go-app-docker); зафиксировать результат
- [x] 1.2 Если api seed отстаёт — догнать seed в coin-api **до** удаления папки (STOP до sync)

## 2. Coordination + docs/ADR

- [x] 2.1 Retarget `pipeline-tekton-alignment` tasks/proposal/design: seed path → `coin-api/internal/gpcontent/seed/`, убрать `coin-gp-content/stacks`
- [x] 2.2 Обновить ADR `gp-embedded-pipeline.md` (секция seed), `build-engine-contract.md`, `coin-ci-runtime.md`
- [x] 2.3 Обновить `architecture.md`, `control-plane.md`, `agent-build-model.md`, `golden-paths.md`, `responsibilities.md`, `jenkins-setup.md`, how-to/runbooks, `coin/README.md`
- [x] 2.4 Синхронизировать main specs `gp-embedded-pipeline` и `runtime-documentation` с delta

## 3. Docker / bootstrap glue

- [x] 3.1 Удалить `docker/scripts/coin-gp-content.sh` и `docker/platform/jenkins/casc-coin-gp-content-build.yaml`
- [x] 3.2 Убрать `make coin-gp-content` из Makefile; вызовы из `bootstrap.sh` / `docker/README.md`

## 4. Удаление папки

- [x] 4.1 Удалить дерево `coin/coin-gp-content/`
- [x] 4.2 `rg coin-gp-content` по активному дереву (исключая `openspec/changes/archive`) — только intentional superseded / historical
- [x] 4.3 Smoke: local API ready + `seed-jenkins-lib` (или эквивалент) без `make coin-gp-content`
