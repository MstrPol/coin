# Разделение ответственности (v2)

## Принцип

- **Проект** — код, `.coin/config.yaml` (identity + credential IDs).
- **Platform** — manifest, GP content, `coin-agent`, coin-api/executor/lib.

## Что управляет разработчик

- Код и зависимости (`go.mod`, …).
- `.coin/config.yaml`: `goldenPath`, `version`, `project.*`, `jenkins.credentials`.
- `Jenkinsfile` — копия `Jenkinsfile.coin` + `@Library('coin-lib@…')`.

## Что управляет Platform

- **coin-api** — composition, catalog policy, resolve, build reports; bootstrap pipeline seed (`internal/gpcontent/seed/`).
- **coin-agent** — universal inbound-agent image (`coin-executor/Dockerfile.agent`).
- **coin-executor** — validate, build engines, publish, report ([CHARTER](../coin-executor/CHARTER.md)).
- **coin-lib** — Jenkins glue only: resolve, pod, credentials, stage dispatch.
- **coin-starters** — product scaffolding + thin Jenkinsfile.
- Platform CI: `coin-executor`, `coin-lib`, `publish-agent`, `seed-jenkins-lib`.

**Superseded:** `coin-jenkins-agents/`, job `agents-build`, GP `scripts/*.sh` в runtime, папка `coin-gp-content/`.

## Граница coin-executor

**В executor:** validate config vs manifest, materialize Containerfile, dispatch `buildkit`/`buildpack`/`dockerfile`, run stages, report.

**Не в executor:** GP publish, semver bump GP release, release notes authoring.

## Что запрещено в проекте

| Запрет | Причина |
|--------|---------|
| `Dockerfile` в репо (go GP) | Managed Containerfile из manifest |
| `template`/`templateVersion` (v1) | Strict v2 only |
| Pin executor/agent/build engine в config | Только в manifest / GP |
| Бизнес-логика сборки в Jenkinsfile/Groovy | coin-executor + GP content |

## Артефакты

| Артефакт | Владелец |
|----------|----------|
| `coin-api`, `coin-executor` | Platform |
| GP pipeline (embedded body + seed defaults) | GP release / `coin-api` seed |
| `coin-agent` image | Platform (`publish-agent`) |
| Jenkins glue (`coinPipeline`) | `coin-lib` (Gitea tag) |
| `.coin/config.yaml` | Команда |
| App OCI image | Команда (registry) |

## Связанные документы

- [config.md](config.md)
- [architecture.md](architecture.md)
- [agent-build-model.md](agent-build-model.md)
