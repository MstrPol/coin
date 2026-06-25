# platform-runtime-line Specification

## Purpose
TBD - created by archiving change gp-profile-metadata-model. Update Purpose after archive.

## Requirements

### Requirement: Platform runtime line configuration

The coin-api SHALL maintain a platform-level pin for **lib** (Jenkins Shared Library) independent of any GP profile.

Agent stack and executor versions SHALL be selected per GP draft (not in platform runtime).

#### Scenario: Read platform runtime

- **WHEN** admin reads platform settings or dedicated runtime endpoint
- **THEN** the response MUST include the current `lib` pin used during manifest resolve

### Requirement: Resolve injects platform lib and GP draft pins

Manifest resolve for a GP release SHALL merge GP draft composition (`agent`, `gp-content`, `branching-model`) with platform `lib` pin to produce a complete manifest.

Executor section SHALL be derived from the selected agent stack.

#### Scenario: Resolve published release

- **WHEN** resolve runs for GP `go-app@1.0.0` with three-pin composition
- **THEN** the resolved manifest MUST include `lib` from platform runtime
- **AND** `runtime` / executor from the pinned agent stack
- **AND** `build` / pipeline from gp-content
- **AND** `branching` from branching-model

#### Scenario: Missing platform lib

- **WHEN** platform `lib` pin is not configured
- **THEN** resolve MUST fail with an explicit error (not silent omission)
