## Why

Текущий job `coin-executor` собирает и публикует только бинарь `coin-executor` в Maven, но не закрывает целевой runtime-поток для local pilot: сборку `coin-agent` образа, публикацию в Docker registry Nexus и регистрацию draft-версии `agent/coin-agent` в платформе. Из-за этого итоговый E2E прогон требует ручного запуска отдельных скриптов и расходится с ожиданием от CI job.

## What Changes

- Переписать `coin-executor/Jenkinsfile` под пайплайн сборки `coin-agent` образа вместо отдельной публикации бинаря.
- Вынести фактическую бизнес-логику сборки/публикации в `scripts/publish-agent.sh`, а в Jenkinsfile оставить orchestration (параметры, stages, credentials binding, вызов скрипта).
- Обеспечить публикацию образа `coin-agent:<version>` в Nexus Docker repository.
- После успешного push автоматически создавать draft-версию `agent/coin-agent@<version>` через coin-api Admin API.
- Сохранить manual gate на promote: job не публикует `agent` в stable, только готовит draft.

## Non-goals

- Не менять контракт GP composition (остаются три pin: `agent`, `gp-content`, `branching-model`).
- Не вводить auto-promote `agent` после регистрации draft.
- Не добавлять бизнес-логику сборки в Jenkins Shared Library (`coin-lib`), кроме Jenkins-only glue.
- Не менять schema/контракт `coin-api` Admin API для agent draft create.

## Capabilities

### New Capabilities
- `coin-executor-agent-pipeline`: Jenkins job `coin-executor` собирает и публикует `coin-agent` образ и регистрирует `agent` draft в платформе.

### Modified Capabilities
- `runtime-agent-registry`: уточняется CI-путь подготовки `agent` draft через job `coin-executor` (build image + push + draft register), без изменения manual promote gate.

## Impact

- Affected code: `coin-executor/Jenkinsfile`, возможные доработки `coin-executor/scripts/publish-agent.sh` и связанных helper scripts.
- Affected systems: Jenkins, Nexus Docker (`coin-docker`), coin-api Admin API (`/v1/admin/components/agent/...`).
- Dependencies: Jenkins credentials (`nexus-docker`, `coin-publisher-api-key`), доступность `docker` daemon в Jenkins environment.
- Постоянные архитектурные инварианты остаются в рамках ADR `docs/adr/coin-ci-runtime.md` и `docs/adr/jenkins-lib-http-nexus.md`.
