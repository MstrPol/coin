## MODIFIED Requirements

### Requirement: Three-pin GP draft composition

GP release composition SHALL содержать ровно три component pin, выбранных оператором:

1. **`agent`** — CI runtime stack (container image with baked `coin-executor`; e.g. `coin-agent`, `agent-30-06`)
2. **`gp-content`** — build policy, Containerfile, schema
3. **`branching-model`** — versioning and publish policy

Standalone `executor` SHALL NOT появляться в GP composition map, registry или resolved manifest. Agent pin является единственным runtime source of truth.

Для local pilot существующая запись GP version/release SHALL также содержать поля publish/cache destination base URL, которые материализуются в секцию `destinations` resolved manifest. Destination fields являются частью GP release identity для расчёта manifest hash, но не являются четвёртым component pin, platform component или отдельной destination model/catalog и SHALL NOT храниться в `gp-content`.

Resolved manifest SHALL быть deterministic materialization из GP release identity, destination fields версии GP и этих трёх pins. `coin-api` SHALL NOT добавлять site-local Jenkins glue fields, credential IDs или synthetic runtime sections, которые не получены из GP release identity, destination fields версии GP, `agent`, `gp-content` или `branching-model`.

#### Scenario: Создание GP draft с тремя pins

- **WHEN** publisher создаёт GP draft с версиями agent, gp-content и branching-model
- **THEN** coin-api MUST сохранить ровно три composition rows
- **AND** MUST NOT валидировать или требовать `executor/coin-executor@{agentVersion}` в component registry

#### Scenario: Отклонение standalone executor в GP draft composition

- **WHEN** publisher пытается создать draft с `executor` как отдельным composition key
- **THEN** coin-api MUST отклонить запрос с invalid composition error

#### Scenario: Resolve материализует runtime только из agent

- **WHEN** GP release resolves для CI
- **THEN** coin-api MUST заполнить `manifest.runtime` из metadata pinned agent version
- **AND** MUST NOT добавлять `manifest.executor`
- **AND** MUST NOT запрашивать component registry для type `executor`

#### Scenario: Resolve возвращает pilot destinations из полей GP version

- **WHEN** GP release `gp-01-07@1.0.0` resolves для CI
- **THEN** manifest MUST содержать `destinations`, материализованные из полей существующей GP version/release record
- **AND** destinations MUST NOT материализоваться из `gp-content`
- **AND** destinations MUST NOT материализоваться из defaults `coin-lib`
- **AND** destinations MUST NOT требовать отдельный platform component или destination catalog

#### Scenario: Resolve возвращает только разрешённые секции

- **WHEN** GP release `gp-01-07@1.0.0` resolves для CI
- **THEN** manifest MUST содержать GP identity fields `goldenPath.name` и `goldenPath.version`
- **AND** MUST содержать `destinations`, материализованные из полей GP version
- **AND** MUST содержать `runtime`, материализованный из `agent` pin
- **AND** MUST содержать `build`, `pipeline`, `validateSchema` и `capabilities`, материализованные из `gp-content` pin
- **AND** MUST содержать `branching`, материализованный из `branching-model` pin
- **AND** MUST сохранять resolve integrity metadata `manifestVersion` и `manifestHash`
- **AND** MUST NOT содержать top-level `credentials`, `lib`, `executor` или любой Jenkins-instance credential ID
