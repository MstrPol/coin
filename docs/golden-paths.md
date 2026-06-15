# Golden paths (Control Plane v2)

Golden path — **semver release** платформы (`go-app@1.0.0`), собранный из registry components + Nexus artifacts.

---

## GP composition — 5 компонентов

| Slot key | Component | Роль |
|----------|-----------|------|
| `jnlp` | `agent/jnlp` | Jenkins inbound agent image |
| `agent` | `agent/{stack}` | CI stack container image |
| `executor` | `executor/coin-executor` | coin-executor binary |
| `lib` | `lib/coin-lib` | Jenkins Shared Library (glue only) |
| `gp-content` | `gp-content/{golden-path}` | scripts, Dockerfile, validate schema |

Профиль создаётся через `POST /v1/admin/golden-paths/profiles` с `{ name, agentStack }` — slots генерирует сервер.

Product Jenkinsfile:

```groovy
@Library('coin-lib@1.0.0') _
coinPipeline()
```

`coin-lib@1.0.0` — стабильный Jenkins adapter для local pilot. Версионируемые части GP приходят из manifest.

---

## GP content

Исходники: [`coin-gp-content/`](../coin-gp-content/).

Publish:

```bash
cd coin-gp-content && ./scripts/publish-content.sh go-app 1.0.0
```

Zip → Nexus `maven-releases` (`coin/gp-content/{name}/{ver}/`) → register `gp-content/go-app@1.0.0` в coin-api.

Manifest resolve включает `jnlp`, `pipeline.stages`, `validateSchema`, `dockerfileTemplate` (без `orchestration.bundle`).

---

## coin-lib

Jenkins glue: [`coin-lib/`](../coin-lib/). Phase 1 — Gitea `coin/coin-lib` tag `1.0.0`; target — Nexus HTTP ZIP (см. ADR `jenkins-lib-http-nexus`).
`coin-lib` рендерит Jenkins stages из `manifest.pipeline.stages`, поэтому разные GP могут иметь разные stage sets.

---

## jnlp registration

`agent/jnlp` публикуется вручную через coin-ui / admin API (`POST /v1/admin/components/agent/jnlp/versions`), не через `agents-build`.

---

## Именование

```
go-app              →  goldenPath name (coin.goldenPath)
1.0.0               →  semver GP release (coin.version)
go                  →  agent stack
go-app              →  gp-content name (совпадает с golden path)
```

---

## Deliverables (product config V1)

Секция `deliverables` в `.coin/config.yaml` описывает **WHAT** — outputs repo. Если секция отсутствует, применяется default `app:image`.

P0 types (должны быть в `gp-content` capabilities): `image`, `liquibase-image`, `artifact(format=zip)`.

Build report содержит `outputs[]` с ref/digest/sha256 per deliverable.

### Roadmap (вне local pilot)

| Приоритет | Тип | Назначение |
|-----------|-----|------------|
| P1 | `helm-chart` | OCI/chart registry publish из GP scripts |
| P1 | `static-assets` | S3/Nexus static site или bundle |
| P1 | `job-image` | CronJob/K8s Job image (отдельно от app image) |
| P2 | `library` | Maven/npm и др. binary artifacts |
| P2 | `terraform-module` | Versioned IaC module zip |
| P2 | `config-bundle` | K8s manifests / config-only zip |

Новый тип добавляется только через `gp-content` capabilities + executor support; product repo объявляет WHAT, platform — HOW.

## Extension policy

- **P0:** declarative `artifact.sources` — только explicit paths, без globs
- **Запрещено:** custom Jenkins stages/actions, override стандартных stage scripts
- **P1 hooks:** executor-owned extension points — design only, не реализованы в local pilot

Повторяющиеся hooks → promote в `gp-content` capabilities, не в product repo.

---

## Связанные документы

- [responsibilities.md](responsibilities.md)
- [ADR jenkins-lib-http-nexus](../.cursor/plans/adr/jenkins-lib-http-nexus.md)
