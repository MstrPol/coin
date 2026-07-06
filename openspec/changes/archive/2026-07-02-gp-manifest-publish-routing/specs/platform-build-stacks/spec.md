## MODIFIED Requirements

### Requirement: Редактор GP content schema v2

Build stack editor SHALL редактировать `content.yaml` schema v2 биективно через упорядоченные карточки секций: engine, build policy, capabilities, pipeline stages и artifacts.

`content.yaml` schema v2 SHALL NOT включать cache registry, cache ref, cache ref template, image registry, Maven repository base или Jenkins credential fields. Эти destination values являются полями GP version, материализуемыми в manifest для pilot, а не build stack policy в gp-content.

#### Scenario: Карточка engine переключает build block

- **WHEN** publisher выбирает engine `dockerfile` в editor
- **THEN** UI MUST показать BYO dockerfile fields (`path`, `imageTarget`, `testTarget`)
- **AND** MUST скрыть buildkit targets и managed containerfile artifact key
- **AND** MUST NOT показывать cacheRefTemplate или cache registry fields

#### Scenario: Карточка engine buildkit

- **WHEN** publisher выбирает engine `buildkit`
- **THEN** UI MUST показать buildkit targets map и managed containerfile artifact editor
- **AND** MUST разрешить `artifact` в capabilities deliverables
- **AND** MUST NOT показывать cacheRefTemplate или cache registry fields

#### Scenario: Save создаёт v2 yaml

- **WHEN** publisher сохраняет draft из editor
- **THEN** persisted `content.yaml` MUST иметь `schemaVersion: 2`
- **AND** MUST NOT содержать `controls` или `pipeline.stages[].when`
- **AND** MUST NOT содержать `cacheRefTemplate`
- **AND** MUST NOT содержать physical publish/cache destination URLs
