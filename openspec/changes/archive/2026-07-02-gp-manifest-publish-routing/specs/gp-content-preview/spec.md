## MODIFIED Requirements

### Requirement: GP content preview API

coin-api SHALL предоставлять `POST /v1/admin/gp-content/preview`, принимающий draft gp-content manifest subset или body `content.yaml` и возвращающий resolved fragments `build` и `pipeline`, validation issues и engine-specific warnings.

GP content preview SHALL NOT возвращать cache registry refs, cache ref templates, image registry prefixes, Maven repository bases или другие physical destination values.

#### Scenario: Preview buildkit draft

- **WHEN** publisher отправляет valid buildkit `content.yaml` v2 draft
- **THEN** coin-api MUST вернуть resolved `build.engine` `buildkit` с targets
- **AND** MUST включить `artifacts.containerfile`, когда engine равен buildkit
- **AND** MUST NOT включать cacheRef или cacheRefTemplate

#### Scenario: Preview BYO dockerfile draft

- **WHEN** publisher отправляет valid dockerfile engine v2 draft с `build.dockerfile.path`
- **THEN** coin-api MUST вернуть resolved `build.dockerfile` без managed `containerfile` content ref
- **AND** MUST предупредить, если `capabilities.deliverables` включает `artifact`
- **AND** MUST NOT включать cacheRef или cacheRefTemplate

#### Scenario: Отклонение v1 content

- **WHEN** publisher отправляет content без `schemaVersion: 2`
- **THEN** preview MUST вернуть validation error
