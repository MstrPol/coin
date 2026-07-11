## REMOVED Requirements

### Requirement: Build stacks catalog

**Reason**: Build stack is not a separate platform entity; pipeline is authored on GP release.

**Migration**: Use GP hub → release detail → Pipeline tab.

### Requirement: GP content schema v3 editor

**Reason**: Editor moved to GP release detail.

**Migration**: See `build-stack-pipeline-ui` delta.

### Requirement: Build stack preview panel

**Reason**: Preview moved to GP release pipeline preview API.

**Migration**: See `gp-embedded-pipeline`.

### Requirement: gp-content from GP release composition

**Reason**: No gp-content composition pin or platform catalog link.

**Migration**: Pipeline tab on same GP release detail page.
