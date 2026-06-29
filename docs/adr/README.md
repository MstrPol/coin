# ADR — Coin

Architecture Decision Records — постоянные архитектурные решения проекта.

Правила: [plan-execution.mdc](../../.cursor/rules/plan-execution.mdc)

## Документы

| ADR | Статус | Файл |
|-----|--------|------|
| **Coin CI runtime** (agent, bootstrap, engines, publish) | accepted | [`coin-ci-runtime.md`](coin-ci-runtime.md) |
| Control Plane v2 — три слоя SoT, hard cut lib/cli | accepted | [`control-plane-v2.md`](control-plane-v2.md) |
| GP branching model | accepted | [`gp-branching-model.md`](gp-branching-model.md) |
| Build engine contract (`build.engine` hard cut) | accepted | [`build-engine-contract.md`](build-engine-contract.md) |
| Jenkins Shared Library + gp-content | accepted | [`jenkins-lib-http-nexus.md`](jenkins-lib-http-nexus.md) |
| Jenkins lib вне control plane | accepted | [`jenkins-lib-outside-platform.md`](jenkins-lib-outside-platform.md) |
| GP Component Package Model (UI-first) | accepted | [`gp-component-package-model.md`](gp-component-package-model.md) |
| GP composition — четыре компонента | superseded | [`gp-composition-four-components.md`](gp-composition-four-components.md) |
| Pipeline bundle layer | superseded | [`gp-pipeline-bundle-layer.md`](gp-pipeline-bundle-layer.md) |
| Corp CI/CD migration standards | accepted (corp target) | [`cicd-corp-migration-standards.md`](cicd-corp-migration-standards.md) |

**Читать первым для CI:** [coin-ci-runtime.md](coin-ci-runtime.md) → [agent-build-model.md](../agent-build-model.md).

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
