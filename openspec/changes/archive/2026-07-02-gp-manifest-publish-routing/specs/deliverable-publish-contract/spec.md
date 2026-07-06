## ADDED Requirements

### Requirement: P0 publish deliverables

GP version SHALL разрешать publish только для P0-набора deliverable types: `image`, `liquibase-image`, `artifact`.

GP/Build Stack SHALL полностью задавать publish deliverables для выбранного Golden Path.

Product `.coin/config.yaml` SHALL NOT объявлять `deliverables`.

GP/Build Stack SHALL NOT объявлять больше одного deliverable каждого type в рамках одного GP.

Product `.coin/config.yaml` SHALL NOT задавать произвольные build/publish commands для deliverables.

#### Scenario: Максимальная P0-комбинация

- **WHEN** GP/Build Stack объявляет один `image`, один `liquibase-image` и один `artifact`
- **THEN** validation MUST принять deliverables GP

#### Scenario: GP задаёт subset

- **WHEN** GP/Build Stack объявляет только один `image`
- **THEN** validation MUST принять deliverables GP

#### Scenario: Deliverables в product config отклоняются

- **WHEN** product `.coin/config.yaml` содержит секцию `deliverables`
- **THEN** validation MUST отклонить config

#### Scenario: Несколько image deliverables в GP отклоняются

- **WHEN** GP/Build Stack объявляет два deliverables с type `image`
- **THEN** validation MUST отклонить GP/Build Stack

#### Scenario: Несколько artifact deliverables в GP отклоняются

- **WHEN** GP/Build Stack объявляет два deliverables с type `artifact`
- **THEN** validation MUST отклонить GP/Build Stack

#### Scenario: Unsupported deliverable type отклоняется

- **WHEN** GP/Build Stack объявляет deliverable type вне P0-набора
- **THEN** validation MUST отклонить GP/Build Stack
