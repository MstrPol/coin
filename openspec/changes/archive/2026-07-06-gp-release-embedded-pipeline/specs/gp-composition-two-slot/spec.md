## REMOVED Requirements

### Requirement: Three-pin GP draft composition

**Reason**: Superseded by `gp-release-two-pin` embedded pipeline model.

**Migration**: See `gp-release-two-pin` spec.

### Requirement: gp-content pinned per GP version not profile

**Reason**: gp-content component eliminated.

**Migration**: Pipeline on GP release body.

### Requirement: Component catalog independence

**Reason**: Only agent and branching-model remain as shared catalog components for GP composition.

**Migration**: Agent and branching-model catalog unchanged.

### Requirement: Explicit component names in draft API

**Reason**: Replaced by two-pin draft API without gpContentName.

**Migration**: See `gp-release-two-pin`.

### Requirement: Profile name may differ from gp-content name

**Reason**: No gp-content; profile name equals pipeline family.

**Migration**: Remove alias profiles (`xxx`, `gp-01-07` → go-app).

### Requirement: Accept draft gp-content in GP draft

**Reason**: Embedded pipeline on GP release replaces gp-content draft pin.

**Migration**: Edit pipeline on GP release detail.

### Requirement: GP draft on draft component pins with promote gate

**Reason**: Promote gate scope reduced to agent + branching-model external pins plus embedded pipeline validity.

**Migration**: See `gp-release-two-pin` promote scenarios.
