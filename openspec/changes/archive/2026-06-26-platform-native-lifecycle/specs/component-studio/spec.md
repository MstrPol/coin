## REMOVED Requirements

### Requirement: Component Studio as primary authoring path

**Reason**: Authoring moved to Platform entity pages under `/platform/build-stacks` and `/platform/branching-models`. Component canary lifecycle removed.

**Migration**: Use Platform catalog create/edit/publish flows. `/studio` routes removed.

### Requirement: Type-aware editors

**Reason**: Editors embedded in Platform entity pages; no standalone Studio hub.

**Migration**: Navigate to `/platform/build-stacks/{name}/{version}/edit` or `/platform/branching-models/{name}/{version}/edit`.

### Requirement: Pilot project selection

**Reason**: Pilot selection belongs to GP canary configuration, not component Studio.

**Migration**: Configure pilot projects via GP canary policy and project canary_mode.

### Requirement: Studio entry from platform catalogs

**Reason**: Platform catalogs provide in-place navigation; Studio route removed.

**Migration**: Use Platform entity detail and edit routes.
