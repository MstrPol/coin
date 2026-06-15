# ADR: GP composition — четыре компонента

**Статус:** accepted  
**Дата:** 2026-06-10

## Контекст

GP release не должен содержать freeform 6-slot composition. Оператор pin'ит платформенный runtime и Jenkins adapter.

## Решение

GP profile и GP release composition состоят из **четырёх slots**:

| Slot key | Component |
|----------|-----------|
| `jnlp` | `agent/jnlp` |
| `agent` | `agent/{stack}` |
| `executor` | `executor/coin-executor` |
| `pipeline-bundle` | `pipeline-bundle/{stack}` |

`CreateGPProfile` принимает `{ name, agentStack }`; slots генерирует сервер.

## Последствия

- UI publish: 4 version pickers
- Запрещён freeform slot editor в coin-ui
- `component_compatibility` rules привязаны к `pipeline-bundle`, не к `pipeline`

## Отклонённые альтернативы

- 6-slot GP (pipeline, validate, dockerfile, orchestration отдельно) — superseded
- 3-slot без pipeline-bundle — не покрывает Jenkins adapter immutability
