## ADDED Requirements

### Requirement: Product config без publish repository

Product `.coin/config.yaml` v2 SHALL хранить только GP pin, identity проекта и project-specific Jenkins glue.

Секция `project` SHALL содержать `name`, `groupId` и `artifactId`.

Секция `project` SHALL NOT содержать `repository`, `imageRepository`, `dockerRepository`, `mavenRepository`, `pypiRepository` или другие destination fields.

Product `.coin/config.yaml` SHALL NOT содержать секцию `deliverables`.

Product `.coin/config.yaml` SHALL NOT содержать build commands, publish commands, pipeline stages, cache refs или registry/repository URLs.

#### Scenario: Валидный product config

- **WHEN** product `.coin/config.yaml` содержит `project.name`, `project.groupId` и `project.artifactId`
- **AND** не содержит `project.repository`
- **AND** не содержит `deliverables`
- **THEN** config validation MUST принять project identity

#### Scenario: Repository в product config отклоняется

- **WHEN** product `.coin/config.yaml` содержит `project.repository`
- **THEN** config validation MUST отклонить config как содержащий destination field

#### Scenario: Deliverables в product config отклоняются

- **WHEN** product `.coin/config.yaml` содержит секцию `deliverables`
- **THEN** config validation MUST отклонить config как содержащий build/publish output definition

#### Scenario: Destination приходит из GP manifest

- **WHEN** executor строит publish destination для проекта
- **THEN** executor MUST использовать `project.name`, `project.groupId` и `project.artifactId` из product config
- **AND** MUST использовать repository/base URL только из `manifest.destinations`

#### Scenario: Deliverables приходят из GP manifest

- **WHEN** executor определяет outputs для build/publish
- **THEN** executor MUST использовать deliverables, материализованные из GP/Build Stack в manifest
- **AND** MUST NOT читать deliverables из product `.coin/config.yaml`
