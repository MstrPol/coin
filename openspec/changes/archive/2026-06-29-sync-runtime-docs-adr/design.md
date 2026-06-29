## Context

Coin прошёл несколько hard cut без полной консолидации документации:

| Область | Код (SoT) | Docs сегодня |
|---------|-----------|--------------|
| GP composition | 3 pins: `agent`, `gp-content`, `branching-model`; executor derived | `architecture.md` — 4 slots + отдельный executor |
| Build runtime | `coin-agent` = inbound-agent + executor + podman + pack + buildkit bins | ADR context описывает stack agents |
| Publish gate | `params.publish` → `COIN_PUBLISH_REQUEST` + `manifest.branching` | Примеры с `pipeline.stages[].when: tag` |
| Superseded | pipeline-bundle, jnlp slot, coin-jenkins-agents | ADR в индексе superseded, но тело без banner |

Operational doc [`docs/agent-build-model.md`](../../docs/agent-build-model.md) наиболее точен, но дублирует и расходится с `control-plane.md` / `architecture.md`.

## Goals / Non-Goals

**Goals:**

- Один канонический ADR **`coin-ci-runtime`** для CI pod, agent image, bootstrap, engines, publish layers.
- Индекс ADR и top-level docs отражают three-pin composition и derived executor.
- Superseded ADR явно помечены; читатель не путает pipeline-bundle с coin-lib.
- Примеры manifest/docs согласованы с branching v2 (без `when: tag` как primary gate).
- OpenSpec specs фиксируют doc requirements для regression при следующих changes.

**Non-Goals:**

- gp-content schema v2, build stack editor, preview API.
- Изменения coin-executor / coin-lib / coin-api behavior.
- Corp fleet rollout, HA, OIDC.
- Полный rewrite всех how-to (только точечные правки при grep legacy).

## Decisions

### D1: Новый ADR `coin-ci-runtime` вместо раздувания `build-engine-contract`

**Решение:** добавить `docs/adr/coin-ci-runtime.md` как operational SoT; `build-engine-contract.md` оставить как decision record про введение `build.engine`, с banner «текущая runtime-модель → coin-ci-runtime».

**Альтернатива:** только amend build-engine-contract — отклонено: файл смешивает мотивацию hard cut и day-2 operations.

### D2: Doc hierarchy

```
docs/adr/coin-ci-runtime.md     ← canonical runtime (agent, bootstrap, engines, publish)
docs/agent-build-model.md       ← operational runbook (E2E, troubleshooting); ссылается на ADR
docs/architecture.md            ← high-level map; ссылается на ADR
docs/control-plane.md           ← manifest/resolve; без дублирования engine tables
docs/adr/build-engine-contract  ← historical decision + link
```

### D3: Superseded ADR — banner в шапке файла

Формат:

```markdown
> **Статус: superseded** (2026-06-10)  
> Заменено: coin-lib + gp-content + [coin-ci-runtime](coin-ci-runtime.md)
```

Не удалять файлы — ссылки из archive/plans могут оставаться.

### D4: Publish gate в примерах

Заменить narrative:

| Было | Стало |
|------|-------|
| `publish when: tag` | stage `publish` всегда в pipeline; skip — `params.publish=false` (coin-lib); deny — branching |
| controls в gp-content | не документировать как активный контракт |

Эталонные `content.yaml` в `coin-gp-content` **не трогаем** в этом change (scope gp-content v2).

### D5: Pilot vs corp — явная таблица в ADR

| Environment | buildkit engine implementation | bootstrap |
|-------------|-------------------------------|-----------|
| local pilot arm64 | podman build (buildctl RUN broken) | podman system service only |
| corp amd64 (roadmap) | buildkitd + buildctl | buildkitd + podman per corp ADR |

Имя engine `buildkit` сохраняется в обоих случаях.

### D6: GP composition в architecture.md

Заменить таблицу 4-slot на 3-pin + примечание: `executor` materialized from `agent`; `lib` — platform pin, не в GP composition map.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Два ADR про runtime путают читателя | Чёткая иерархия D2; build-engine-contract → link only |
| Grep находит legacy в 20+ how-to | tasks: targeted grep + fix high-traffic docs only |
| content.yaml эталоны с `when: tag` противоречат docs | Явно в ADR: «reference stacks — cleanup в gp-content-schema-v2» |
| cicd-corp-migration-standards vs pilot podman | Секция «pilot exception» со ссылкой на corp ADR |

## Migration Plan

1. Написать `coin-ci-runtime.md`.
2. Amend superseded banners + `adr/README.md`.
3. Patch `architecture.md`, `control-plane.md`, `agent-build-model.md` (dedupe tables → link ADR).
4. Grep legacy terms; fix `docs/README.md`, `how-to/troubleshoot-ci.md` при необходимости.
5. Archive change → baseline `runtime-documentation` spec.

Rollback: docs-only revert; без runtime impact.

## Open Questions

| # | Вопрос | Статус | Варианты |
|---|--------|--------|----------|
| Q1 | Переименовать `gp-composition-two-slot` spec id? | ✅ defer | Оставить id; обновить Purpose при archive gp-content change |
| Q2 | Включить `docs/adr/gp-branching-model.md` merge в coin-ci-runtime? | ✅ A | Оставить отдельным ADR; cross-link publish policy |
