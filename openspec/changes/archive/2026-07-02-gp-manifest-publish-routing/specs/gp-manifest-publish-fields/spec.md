## ADDED Requirements

### Requirement: Поля publish destinations в GP manifest для pilot

Существующая запись GP version/release SHALL содержать pilot-поля publish destinations для outputs build path.

Resolved GP manifest SHALL включать top-level `destinations`, материализованный из этих полей GP version/release.

Секция `destinations` SHALL содержать `imageRegistryPrefix`, `buildCacheEnabled` и `artifactRepositoryBase`.

Секция `destinations` SHALL NOT содержать отдельный `cacheRegistryPrefix`; build cache ref для pilot строится от `imageRegistryPrefix`.

Секция `destinations` SHALL NOT содержать `packageRepositories` в P0.

Секция `destinations` SHALL входить в canonical manifest JSON и участвовать в расчёте `manifestHash`.

Секция `destinations` SHALL NOT содержать Jenkins credential IDs, usernames, passwords, tokens или secret references.

#### Scenario: Resolve возвращает destinations

- **WHEN** product CI resolves GP `gp-01-07@1.0.0`
- **THEN** возвращённый manifest MUST содержать `destinations.imageRegistryPrefix`
- **AND** MUST содержать `destinations.buildCacheEnabled`
- **AND** MUST содержать `destinations.artifactRepositoryBase`
- **AND** MUST сохранять `manifestHash`, рассчитанный по canonical document с учётом `destinations`

#### Scenario: Fallback manifest содержит destinations

- **WHEN** `coin-api` закешировал resolved GP manifest blob в Nexus
- **AND** product CI позднее выполняет resolve через Nexus fallback path
- **THEN** manifest blob MUST содержать ту же секцию `destinations`, что и primary API resolve result
- **AND** fallback path MUST NOT требовать Jenkins-managed destination config

#### Scenario: Destinations не содержат credentials

- **WHEN** manifest содержит pilot destination URLs
- **THEN** manifest MUST NOT содержать top-level секцию `credentials`
- **AND** `destinations` MUST NOT содержать Jenkins credential ID fields
- **AND** `destinations` MUST NOT содержать secret values

### Requirement: Runtime refs из destinations и project identity

`coin-executor` SHALL строить concrete image, cache и artifact publish destinations из manifest `destinations` и project identity в product `.coin/config.yaml`.

`coin-executor` SHALL брать список publish deliverables из GP/Build Stack, материализованный в manifest, а не из product `.coin/config.yaml`.

App image refs SHALL использовать `destinations.imageRegistryPrefix`, `project.groupId`, `project.artifactId` и `project.name`.

Liquibase image refs SHALL использовать `destinations.imageRegistryPrefix`, `project.groupId`, `project.artifactId` и `project.name` с suffix `-liquibase`.

Build cache refs SHALL использовать `destinations.imageRegistryPrefix`, `project.groupId`, `project.artifactId`, `project.name` и suffix `-cache`.

Artifact publish destinations SHALL использовать `destinations.artifactRepositoryBase` для repository URL, а `project.groupId` и `project.artifactId` остаются artifact coordinates.

#### Scenario: Buildkit cache ref

- **WHEN** executor запускает buildkit stage для проекта `demo-go-app`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest destinations содержат `imageRegistryPrefix: docker-dev.registry.domain.ru`
- **AND** manifest destinations содержат `buildCacheEnabled: true`
- **THEN** executor MUST использовать cache ref `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-cache`

#### Scenario: Image publish ref

- **WHEN** executor собирает image deliverable для проекта `demo-go-app`
- **AND** manifest содержит GP deliverable type `image`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest destinations содержат `imageRegistryPrefix: docker-dev.registry.domain.ru`
- **THEN** executor MUST сформировать image ref внутри `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app`

#### Scenario: Liquibase image publish ref

- **WHEN** executor собирает liquibase-image deliverable для проекта `demo-go-app`
- **AND** manifest содержит GP deliverable type `liquibase-image`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest destinations содержат `imageRegistryPrefix: docker-dev.registry.domain.ru`
- **THEN** executor MUST сформировать image ref внутри `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-liquibase`

#### Scenario: Build cache выключен в GP

- **WHEN** manifest destinations содержат `buildCacheEnabled: false`
- **THEN** executor MUST NOT передавать registry cache options в build engine

#### Scenario: Artifact repository URL

- **WHEN** executor публикует artifact deliverable
- **AND** manifest содержит GP deliverable type `artifact`
- **AND** manifest destinations содержат `artifactRepositoryBase: http://nexus:8081/repository/maven-releases`
- **THEN** executor MUST использовать repository URL `http://nexus:8081/repository/maven-releases`
