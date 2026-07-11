## REMOVED Requirements

### Requirement: Platform settings without runtime lib pin

**Reason**: Platform settings UI and API removed; Nexus integration is configured via coin-api environment variables and docker compose, not operator console.

**Migration**: Configure `NEXUS_URL`, `NEXUS_MAVEN_RELEASES`, `NEXUS_MAVEN_SNAPSHOTS`, `NEXUS_ADMIN_USER`, and `NEXUS_ADMIN_PASSWORD` in docker `.env` / deployment manifests. Former `/platform-settings` bookmarks redirect to `/audit`.

## ADDED Requirements

### Requirement: Legacy platform settings redirect

The coin-ui SHALL redirect former platform settings bookmarks.

#### Scenario: Redirect platform settings URL

- **WHEN** user navigates to `/platform-settings`
- **THEN** the UI MUST redirect to `/audit`

### Requirement: Admin navigation without platform settings

The Admin sidebar group SHALL NOT include a Platform settings entry.

#### Scenario: Admin nav items

- **WHEN** admin views the Admin group in the sidebar
- **THEN** the UI MUST include Audit (`/audit`)
- **AND** MUST NOT include Platform settings or Nexus configuration
