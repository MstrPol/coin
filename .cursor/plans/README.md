# Планы Coin

> **Только здесь:** `coin/.cursor/plans/` — не в `~/.cursor/plans/`.

**Активный plan:** [coin-ui-operator-console.plan.md](coin-ui-operator-console.plan.md) — discovery-driven UI walkthrough

## Модель

Вопросы по coin-ui → проверка стенда → подтверждённые косяки → тикеты **UI-XX** в активном plan.

Стенд: http://localhost:8091 (`make coin-ui-up`), ключ `dev-local-admin-key`.

## Правила

| Файл | Зачем |
|------|-------|
| [plan-execution.mdc](../rules/plan-execution.mdc) | Прогресс, STOP, эскалация |
| [coin-project-gates.mdc](../rules/coin-project-gates.mdc) | Corp gate, SoT |

## Файлы

| Plan | Статус |
|------|--------|
| [coin-ui-operator-console.plan.md](coin-ui-operator-console.plan.md) | **active** |
| [platform-native-jenkins.plan.md](platform-native-jenkins.plan.md) | superseded |

## Завершённые (удалены)

- **platform-first-delivery** (PF-00…PF-25) — MVP-1…3 закрыты; plan удалён после completion

Corp-only follow-up: prod repo extract по [prod-repo-split.md](../../docs/runbooks/prod-repo-split.md).
