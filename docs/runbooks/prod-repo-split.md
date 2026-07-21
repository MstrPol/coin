# Prod repo split (P4-03)

> **⚠️ Corp gate:** выполнять только после доступа в corp-сеть и завершения Wave 3 prep (P3-04).  
> **Ticket:** P4-03

## Цель

Split monorepo dev layout → production Gitea repos:

- `coin/coin-api`
- `coin/coin-executor`
- `coin/coin-ui`
- `coin/coin-lib`
- `coin/coin-starters`

> **Removed:** отдельный corp/local repo `coin-gp-content` — bootstrap seed в `coin-api/internal/gpcontent/seed/`; authoring на GP release.

**Superseded:** `coin/coin-jenkins-agents` — заменён universal `coin-agent` из `coin-executor`.

## Prerequisites

- [ ] Wave 3 runbook reviewed ([wave-3-migration.md](wave-3-migration.md))
- [ ] Fleet scanner green на corp Gitea
- [ ] coin-api HA (P1-05) или accepted SPOF window
- [ ] OIDC prod ([coin-ui user guide](../coin-ui-user-guide.md) RBAC section)
- [ ] Jenkins corp использует v2 только (audit: нет v1 Jenkinsfile в product repos)

## Monorepo layout (PF-16 — dev prep)

В integration monorepo split **уже отражён**:

| Corp repo | Monorepo path | Local Gitea |
|-----------|---------------|-------------|
| `coin/coin-lib` | `coin-lib/` | `make coin-lib` |
| `coin/coin-starters` | `coin-starters/` | `make coin-starters` |
| — | *(removed)* | `coin-jenkins-agents/` superseded by `coin-agent` |

## Extract prod repos

Для каждого компонента:

| Repo | Source path | CI |
|------|-------------|-----|
| `coin/coin-api` | `coin-api/` | Jenkinsfile in repo, image → registry |
| `coin/coin-executor` | `coin-executor/` | publish to Nexus maven-releases |
| `coin/coin-ui` | `coin-ui/` | image + static nginx |
| `coin/coin-lib` | `coin-lib/` | Gitea tag + Shared Library |
| `coin/coin-starters` | `coin-starters/` | product scaffolding |

### Checklist per repo

- [ ] `git filter-repo` или subtree split из monorepo
- [ ] Перенести migrations, openapi, Dockerfile
- [ ] Jenkins multibranch / platform job
- [ ] VERSION / semver policy ([branching-models.md](../how-to/branching-models.md))
- [ ] Update import paths if module rename (keep `coin.local/coin-api` or corp module path)
- [ ] Deploy manifest (k8s) — out of scope local pilot

| Monorepo after split | — |
|------|-----|
| `coin-jenkins-agents/` | **удалён** — `coin-agent` в `coin-executor/Dockerfile.agent` |
| `samples/*` | E2E в monorepo или отдельный `coin/samples` repo |

## Verify

- [ ] corp Jenkins build demo-go-app against prod coin-api URL
- [ ] coin-ui prod SSO login
- [ ] Publish GP release через prod coin-ui (publisher role)
- [ ] No references to monorepo paths in corp Jenkins env

## Rollback

**Откат на config v1 / Shared Library не поддерживается.** При проблемах split — fix forward в prod repos.

## Local pilot (сейчас)

Monorepo остаётся единственным dev layout. Этот runbook — подготовка; команды **не выполняют** split на local Gitea.

## Связанные документы

- [onboarding-15min.md](../how-to/onboarding-15min.md)
- [docs/README.md](../README.md)
