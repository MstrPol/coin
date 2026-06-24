# ADR — Coin

Architecture Decision Records — постоянные архитектурные решения проекта.

Правила: [plan-execution.mdc](../../.cursor/rules/plan-execution.mdc)

## Документы

| ADR | Статус | Файл |
|-----|--------|------|
| Control Plane v2 — три слоя SoT, hard cut lib/cli | accepted | [`control-plane-v2.md`](control-plane-v2.md) |
| GP composition — четыре компонента | superseded | [`gp-composition-four-components.md`](gp-composition-four-components.md) |
| Pipeline bundle layer | superseded | [`gp-pipeline-bundle-layer.md`](gp-pipeline-bundle-layer.md) |
| Jenkins Shared Library + gp-content | accepted | [`jenkins-lib-http-nexus.md`](jenkins-lib-http-nexus.md) |
| Build engine contract | accepted | [`build-engine-contract.md`](build-engine-contract.md) |
| GP Component Package Model (UI-first) | accepted | [`gp-component-package-model.md`](gp-component-package-model.md) |

## Формат нового ADR

1. Контекст
2. Решение
3. Последствия
4. Отклонённые альтернативы

Размещение: `docs/adr/<topic>.md`

## Связанные документы

- Активные changes: [`openspec/changes/`](../../openspec/changes/)
- Планирование: [`docs/planning.md`](../planning.md)
- Docs: [`docs/control-plane.md`](../control-plane.md)
