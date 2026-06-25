## REMOVED Requirements

### Requirement: Platform runtime line configuration

**Reason**: Jenkins Shared Library is outside coin-api control plane; lib version is managed by Jenkins org configuration, not platform operator settings.

**Migration**: Remove `platform_settings.runtime` and Platform settings lib pin UI. coin-lib publishes to Nexus only.

### Requirement: Resolve injects platform lib and GP draft pins

**Reason**: Resolved manifest serves coin-executor; lib bootstrap is independent of manifest resolve.

**Migration**: Remove `lib` from manifest schema and resolve merge logic. Keep executor materialization from agent stack (see `gp-composition-two-slot`).
