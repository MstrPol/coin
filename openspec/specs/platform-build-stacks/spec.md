# platform-build-stacks Specification

## Purpose

Superseded: build stack is not a separate platform entity; pipeline is authored on GP release detail. Retained for delta traceability only.

## Requirements

### Requirement: Build stacks catalog removed

Platform build stacks catalog and gp-content editor SHALL NOT be exposed. Pipeline authoring SHALL occur on GP release detail only.

#### Scenario: Pipeline authoring on GP release detail

- **WHEN** publisher needs to edit pipeline for profile `go-app`
- **THEN** the UI MUST offer pipeline editor on GP release detail for draft releases
- **AND** MUST NOT require navigation to `/platform/build-stacks`
