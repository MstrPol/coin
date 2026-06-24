## MODIFIED Requirements

### Requirement: Component Studio as primary authoring path

The coin-ui SHALL provide Component Studio as the primary path for enabling team to create and publish platform components without git or shell scripts.

#### Scenario: Create draft component

- **WHEN** enabling team creates a new component version in Component Studio
- **THEN** the UI MUST save it as `draft` in PostgreSQL and MUST NOT require a git commit or publish script

#### Scenario: Validate before publish

- **WHEN** enabling team clicks Validate
- **THEN** the UI MUST run server-side schema validation and show errors before allowing publish to canary

#### Scenario: Publish to canary

- **WHEN** enabling team publishes a validated draft to canary for `branching-model`
- **THEN** the UI MUST call Admin API to register the package in PostgreSQL only (artifact bodies + content_ref manifest subset), set state to `canary`, and MUST NOT upload immutable package files to Nexus

#### Scenario: Promote to stable

- **WHEN** health gate passes for pilot projects on canary line and enabling team promotes `branching-model` to stable
- **THEN** the UI MUST upload the immutable package to Nexus, update `content_ref` v2 with `package.url` and digest, and set component state to `published` in one flow
