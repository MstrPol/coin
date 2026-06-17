# Планы Coin

> **Только здесь:** `coin/.cursor/plans/` — не в `~/.cursor/plans/`.

**Активный plan:** [build-engine-model.plan.md](build-engine-model.plan.md) — hard cut build-модели: `build.engine`, universal coin-agent, buildpack/buildkit/dockerfile.

## Модель

Реализация по тикетам активного plan-файла. Старые GP/pipeline-bundle решения считаются superseded. `coin-lib` остается glue-only.

Стенд: http://localhost:8091 (`make coin-ui-up`), ключ `dev-local-admin-key`.

## Правила

| Файл | Зачем |
|------|-------|
| [plan-execution.mdc](../rules/plan-execution.mdc) | Прогресс, STOP, эскалация |
| [coin-project-gates.mdc](../rules/coin-project-gates.mdc) | Corp gate, SoT |

## Файлы

| Plan | Статус |
|------|--------|
| [build-engine-model.plan.md](build-engine-model.plan.md) | **active** |
| coin-lib.plan.md | completed, removed from active plans |
| gp-four-component-model.plan.md | superseded, removed from active plans |
| [platform-native-jenkins.plan.md](platform-native-jenkins.plan.md) | superseded |

## Завершённые (удалены)

- **jenkins-lib-nexus** — миграция на coin-lib + gp-content
- **thin-jenkinsfile-coin-lib** — product Jenkinsfile в 2 строки
- **coin-lib** — гигиена Shared Library, layered config, logging, build parameter
- **coin-ui-operator-console** (UI-00…UI-09) — operator console, 2026-06
- **platform-first-delivery** (PF-00…PF-25) — MVP-1…3 закрыты

Corp-only follow-up: prod repo extract по [prod-repo-split.md](../../docs/runbooks/prod-repo-split.md).
