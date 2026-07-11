## ADDED Requirements

### Requirement: Build cache из manifest destinations

Build cache destination SHALL вычисляться `coin-executor` из manifest `destinations.imageRegistryPrefix`, product `project.groupId`, `project.artifactId`, `project.name` и suffix `-cache`.

Resolved manifest SHALL NOT требовать `build.buildkit.cacheRef` или `build.dockerfile.cacheRef` для runtime cache configuration в executor.

Если `manifest.destinations.buildCacheEnabled` равен `false`, executor SHALL NOT использовать registry cache для build engine.

#### Scenario: Buildkit cache из destinations

- **WHEN** executor запускает buildkit engine для проекта `demo-go-app`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest содержит `destinations.imageRegistryPrefix: docker-dev.registry.domain.ru`
- **AND** manifest содержит `destinations.buildCacheEnabled: true`
- **THEN** executor MUST передать `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-cache` как BuildKit registry cache ref

#### Scenario: Dockerfile engine cache из destinations

- **WHEN** executor запускает dockerfile engine для проекта `demo-go-app`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest содержит `destinations.imageRegistryPrefix: docker-dev.registry.domain.ru`
- **AND** manifest содержит `destinations.buildCacheEnabled: true`
- **THEN** executor MUST передать `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-cache` как BuildKit registry cache ref

#### Scenario: Cache выключен

- **WHEN** manifest содержит `destinations.buildCacheEnabled: false`
- **THEN** executor MUST NOT передавать registry cache ref в build engine

### Requirement: Image publish из manifest destinations

App image deliverable refs SHALL вычисляться из manifest `destinations.imageRegistryPrefix`, product `project.groupId`, `project.artifactId`, `project.name` и resolved Coin version/tag.

Liquibase image deliverable refs SHALL вычисляться из manifest `destinations.imageRegistryPrefix`, product `project.groupId`, `project.artifactId`, `project.name` с suffix `-liquibase` и resolved Coin version/tag.

`coin-lib` SHALL NOT предоставлять image registry prefix как source of truth для image destination.

#### Scenario: Build output image ref

- **WHEN** executor собирает проект `demo-go-app`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest содержит `destinations.imageRegistryPrefix: docker-dev.registry.domain.ru`
- **THEN** executor MUST записать image output ref внутри `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app`

#### Scenario: Build output liquibase image ref

- **WHEN** executor собирает liquibase-image для проекта `demo-go-app`
- **AND** project config содержит `groupId: com.example.team` и `artifactId: demo-go-app`
- **AND** manifest содержит `destinations.imageRegistryPrefix: docker-dev.registry.domain.ru`
- **THEN** executor MUST записать image output ref внутри `docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-liquibase`

#### Scenario: Publish использует build output ref

- **WHEN** executor запускает publish после build
- **THEN** executor MUST опубликовать image ref, записанный во время build
- **AND** MUST NOT заменять registry prefix из defaults `coin-lib`
