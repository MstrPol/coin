# platform-build-stacks Specification

## Purpose

Platform catalog and GP release integration for gp-content build stacks.

## Requirements

### Requirement: Build stacks catalog

The coin-ui SHALL provide a Platform → Build stacks catalog for all `gp-content` components.

#### Scenario: List gp-content stacks

- **WHEN** enabling team opens `/platform/build-stacks`
- **THEN** the UI MUST list all `gp-content` components from the registry with versions and lifecycle status
- **AND** MUST NOT filter or align stacks by GP profile name

#### Scenario: Open stack in Studio

- **WHEN** enabling team selects a gp-content version
- **THEN** the UI MUST link to Component Studio at `/studio/gp-content/{name}/{version}` when publisher role is present

### Requirement: gp-content from GP release composition

When viewing a GP release, the UI SHALL surface the pinned gp-content from that release's composition — not from the profile entity.

#### Scenario: Studio link from release composition

- **WHEN** enabling team views release detail and the composition includes `gp-content/go-app@1.0.0`
- **THEN** the UI MUST offer a link to Studio or Platform build stacks for that component version
- **AND** MUST NOT link to a profile-level Build stack hub tab
