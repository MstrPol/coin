# Разделение ответственности (v2)

## Принцип

- **Проект** — код, `.coin/config.yaml` (identity + credential IDs).
- **Platform** — manifest, content, agent images, coin-api/executor.

## Что управляет разработчик

- Код и зависимости (`go.mod`, …).
- `.coin/config.yaml`: `goldenPath`, `version`, `project.*`, `jenkins.credentials`.
- `Jenkinsfile` — копия `Jenkinsfile.coin` (или symlink policy команды).

## Что управляет Platform

- **coin-api** — composition, catalog policy, resolve.
- **GP content** — coin-api + PostgreSQL + Nexus (не git).
- **Agent images** — `coin-jenkins-agents/`.
- **coin-executor** — bounded runtime (см. [CHARTER](../coin-executor/CHARTER.md)).
- **Agent images** — `coin-jenkins-agents/`.
- **Universal Jenkinsfile** — `starters/Jenkinsfile.coin`.
- Platform CI: `coin-executor`, `coin-gp-content`, `coin-lib`, `agents-build`.
- **coin-lib** — Jenkins Shared Library (glue only: resolve, pod, credentials, stage dispatch).
- **coin-gp-content** — scripts, Dockerfile, schema per golden path (Nexus + coin-api).

## Граница coin-executor

**В executor:** validate config vs manifest, materialize scripts, run stages, report.

**Не в executor:** GP publish, version bump, Dockerfile engine, release notes.

## Что запрещено в проекте

| Запрет | Причина |
|--------|---------|
| `Dockerfile` в репо | Managed template из manifest |
| `template`/`templateVersion` (v1) | Strict v2 only |
| Pin executor/agent в config | Только в manifest |
| v1 Shared Library pipeline | Hard cut v2 |

## Артефакты

| Артефакт | Владелец |
|----------|----------|
| `coin-api`, `coin-executor` | Platform |
| GP content (scripts, Dockerfile, schema) | `coin-gp-content` CI → Nexus + `gp_artifact_bodies` |
| Jenkins glue (`coinPipeline`) | `coin-lib` (Gitea tag 1.0.0 phase 1, Nexus HTTP ZIP target) |
| `.coin/config.yaml` | Команда |
| App OCI image | Команда (registry) |

## Связанные документы

- [config.md](config.md)
- [architecture.md](architecture.md)
- [agent-build-model.md](agent-build-model.md)
