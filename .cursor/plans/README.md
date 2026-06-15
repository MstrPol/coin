# Планы Coin

> **Только здесь:** `coin/.cursor/plans/` — не в `~/.cursor/plans/`.

**Активный plan:** [coin-lib.plan.md](coin-lib.plan.md) — гигиена Shared Library, layered config (lib → GP → project), cleanup `src/coin/`.

## Модель

Реализация по тикетам активного plan-файла. Старые GP/pipeline-bundle решения считаются superseded новым `coin-lib` подходом.

Стенд: http://localhost:8091 (`make coin-ui-up`), ключ `dev-local-admin-key`.

## Правила

| Файл | Зачем |
|------|-------|
| [plan-execution.mdc](../rules/plan-execution.mdc) | Прогресс, STOP, эскалация |
| [coin-project-gates.mdc](../rules/coin-project-gates.mdc) | Corp gate, SoT |

## Файлы

| Plan | Статус |
|------|--------|
| [coin-lib.plan.md](coin-lib.plan.md) | **active** |
| gp-four-component-model.plan.md | superseded, removed from active plans |
| [platform-native-jenkins.plan.md](platform-native-jenkins.plan.md) | superseded |

## Завершённые (удалены)

- **jenkins-lib-nexus** — миграция на coin-lib + gp-content
- **thin-jenkinsfile-coin-lib** — product Jenkinsfile в 2 строки
- **coin-ui-operator-console** (UI-00…UI-09) — operator console, 2026-06
- **platform-first-delivery** (PF-00…PF-25) — MVP-1…3 закрыты

Corp-only follow-up: prod repo extract по [prod-repo-split.md](../../docs/runbooks/prod-repo-split.md).
