## Context

После embedded pipeline:

- GP composition = 2 pins (`agent`, `branching-model`); pipeline — body GP release.
- Bootstrap defaults уже `go:embed` в `coin-api/internal/gpcontent/seed/` (`pipelines/go-app.yaml`, `go-app-docker.yaml`, schemas, Containerfile).
- Папка `coin/coin-gp-content/` почти байт-в-байт дублирует seed (`kind: gp-content` vs `golden-path`) + deprecated publish + local Gitea/Jenkins mirror.

ADR [gp-embedded-pipeline](../../docs/adr/gp-embedded-pipeline.md) ещё называет папку «git seed source» — это устарело относительно фактического api embed.

Активный change `pipeline-tekton-alignment` ссылается на stacks в папке — риск dual v4.

## Goals / Non-Goals

**Goals:**

- Один seed SoT: coin-api embed.
- Убрать папку и local CI glue без регресса pilot seed/E2E.
- Поправить docs/ADR и retarget tekton-alignment tasks.

**Non-Goals:**

- UI cleanup build-stacks / gp-content editor.
- Реализация pipeline v4.
- Перенос seed из api в `docker/testdata` (api embed предпочтительнее для bootstrap).

## Decisions

### D1. SoT seed = `coin-api/internal/gpcontent/seed/` (не testdata)

| | |
|--|--|
| **Выбор** | Не копировать stacks в `docker/testdata/`. Оставить (и документировать) api embed. |
| **Почему** | Seed уже используется bootstrap/GP create; `go:embed` — единый путь с runtime coin-api. Как branching fixtures в testdata не нужен — content тяжелее и уже в api. |
| **Альтернатива** | Testdata + удалить api seed — ломает текущий bootstrap, больший blast radius. |

**Safety gate:** перед `rm` сравнить `coin-gp-content/stacks/{go-app,go-app-docker}/content.yaml` с `seed/pipelines/*.yaml` (допустимы только `kind` и мелкий drift). Если api seed отстаёт — сначала догнать **в coin-api** (отдельный минимальный PR / задача вне `allowedEditRoots` этого change, или координация), потом удалять папку.

> **Примечание scope:** реализация правок в `coin-api` — соседний репозиторий workspace. В `coin` change: verify + docs; при отставании seed — STOP и эскалация / отдельный commit в coin-api до удаления папки.

### D2. Удалить local Gitea/Jenkins job целиком

| | |
|--|--|
| **Выбор** | Убрать `make coin-gp-content`, `coin-gp-content.sh`, CASC job, строки bootstrap. |
| **Почему** | Job только зеркалил папку; publish-content deprecated; pilot не зависит от Gitea `coin/coin-gp-content` для resolve. |
| **Альтернатива** | Оставить пустой repo — шум. |

### D3. Retarget `pipeline-tekton-alignment` в том же PR-окне

| | |
|--|--|
| **Выбор** | В tasks этого change: правка `pipeline-tekton-alignment/tasks.md` (+ proposal/design impact row) — seed path = `coin-api/internal/gpcontent/seed/`, не `coin-gp-content/stacks`. |
| **Почему** | Иначе active change снова воссоздаст dual path. |
| **Альтернатива** | Ждать конца tekton — дольше dual path. |

### D4. Docs / ADR batch

Обновить как минимум:

- `docs/adr/gp-embedded-pipeline.md` — секция coin-gp-content → api seed
- `docs/adr/build-engine-contract.md`, `coin-ci-runtime.md` — SoT path
- `docs/architecture.md`, `control-plane.md`, `agent-build-model.md`, `golden-paths.md`, `responsibilities.md`, `jenkins-setup.md`, docker README, `coin/README.md`
- Runbooks `prod-repo-split.md` — пометить папку removed / superseded

How-to `publish-gp-release.md` уже ближе к правде — подчистить deprecated publish-content path.

### D5. UI build-stacks — out of scope

Оставить как follow-up. Safe remove папки не требует UI delete; мёртвый UI не читает `coin-gp-content/` с диска.

## Risks / Trade-offs

| Риск | Митигация |
|------|-----------|
| Api seed drift vs stacks | Diff gate до удаления; догнать seed при расхождении |
| `pipeline-tekton-alignment` снова пишет в удалённый path | Retarget tasks в этом change |
| Docs 404 | `rg coin-gp-content` вне archive после удаления |
| Кто-то зовёт `make coin-gp-content` | Убрать target; bootstrap comment |
| Corp runbook prod-repo-split | Текст «removed / seed in coin-api»; не corp rollout |

## Migration Plan

1. Diff stacks ↔ api seed (go-app, go-app-docker).
2. Если нужно — sync seed в coin-api (вне/до удаления).
3. Retarget `pipeline-tekton-alignment` artifacts.
4. Обновить docs/ADR + docker glue.
5. Удалить `coin-gp-content/`.
6. `rg coin-gp-content` (исключая archive) — только intentional superseded mentions.
7. Smoke: `seed-jenkins-lib` / ready API при доступном stack.

**Rollback:** restore папку из git; вернуть make target. PG/Nexus не затрагиваются.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| — | _(нет blocking)_ | — | — | SoT = api seed; UI cleanup — follow-up |
