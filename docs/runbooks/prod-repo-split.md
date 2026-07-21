# Prod repo split (P4-03)

> **⚠️ Corp gate:** выполнять только после доступа в corp-сеть и завершения Wave 3 prep (P3-04).  
> **Ticket:** P4-03  
> **Local layout сейчас:** [workspace-layout.md](../workspace-layout.md) — sibling checkouts; этот runbook **не** меняет local pilot.

## Цель

Из integration workspace (sibling folders) извлечь **production Gitea repos**:

| Corp Gitea | Source (workspace) | Назначение |
|------------|-------------------|------------|
| `coin/coin-api` | `coin-api/` | Control plane API |
| `coin/coin-executor` | `coin-executor/` | CLI + coin-agent image |
| `coin/coin-ui` | `coin-ui/` | Admin / Platform SPA |
| `coin/coin-lib` | `coin-lib/` | Jenkins Shared Library |
| `coin/coin-starters` | `coin/coin-starters/` | Product scaffolding |

**Не extract targets (удалены / не нужны как prod content repos):**

- `coin-gp-content` — seed в `coin-api/internal/gpcontent/seed/`; authoring на GP release
- `coin-branching-models` — Platform + `coin/docker/testdata/`
- `coin-jenkins-agents` — superseded `coin-agent`

Meta `coin/docs`, `coin/openspec`, `coin/docker` — docs/pilot tooling; стратегия corp docs repo — вне обязательного P4-03 app split.

## Prerequisites

- [ ] Wave 3 runbook reviewed ([wave-3-migration.md](wave-3-migration.md))
- [ ] Fleet / build-reports green на corp Gitea (по актуальному acceptance)
- [ ] coin-api HA (P1-05) или accepted SPOF window
- [ ] OIDC prod ([coin-ui user guide](../coin-ui-user-guide.md) RBAC)
- [ ] Jenkins corp — product config v2 only

## Local pilot (сейчас)

Monorepo/workspace layout уже **разделён по каталогам** (см. [workspace-layout](../workspace-layout.md)).  
Команды **не выполняют** corp `git filter-repo` на local Gitea. Local Gitea зеркала (`make coin-lib`, …) — только pilot.

## Extract prod repos

### Checklist per repo

- [ ] `git filter-repo` или subtree split из соответствующего sibling path
- [ ] Перенести migrations / openapi / Dockerfile (для api/ui/executor)
- [ ] Jenkins multibranch / platform job
- [ ] VERSION / semver policy ([branching-models.md](../how-to/branching-models.md))
- [ ] Module path (`coin.local/...` или corp module path)
- [ ] Deploy manifest (k8s) — out of scope local pilot

| После split | |
|-------------|--|
| `coin/samples` или corp `coin/samples` | E2E demos |
| `coin/docker` | остаётся local-only |

## Verify

- [ ] corp Jenkins build `demo-go-app` against prod coin-api URL
- [ ] coin-ui prod SSO login
- [ ] Publish GP release через prod coin-ui (publisher+)
- [ ] No references to local monorepo-only paths in corp Jenkins env

## Rollback

**Откат на config v1 не поддерживается.** При проблемах split — fix forward в prod repos.

## Связанные документы

- [workspace-layout.md](../workspace-layout.md)
- [architecture.md](../architecture.md)
- [onboarding-15min.md](../how-to/onboarding-15min.md)
- [docs/README.md](../README.md)
