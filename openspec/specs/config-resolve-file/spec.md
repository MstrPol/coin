# config-resolve-file Specification

## Purpose

Product `.coin/config.yaml` resolve modes: default remote (coin-api → Nexus) and local `file` for platform/sample debugging.

## Requirements

### Requirement: Product config resolve mode

Product `.coin/config.yaml` v2 SHALL support optional `coin.resolve` with values `remote` or `file`. When `coin.resolve` is omitted, behavior MUST be identical to `remote`. `coin.goldenPath` and `coin.version` MUST remain required for both modes.

#### Scenario: Default remote resolve

- **WHEN** product config omits `coin.resolve`
- **THEN** coin-lib MUST resolve manifest via coin-api with Nexus fallback

#### Scenario: File resolve requires goldenPath and version

- **WHEN** product config sets `coin.resolve: file` without `coin.goldenPath` or `coin.version`
- **THEN** config validation MUST fail

### Requirement: Manifest file path for file resolve

When `coin.resolve` is `file`, coin-lib SHALL load the resolved manifest from `coin.manifestFile` if set, otherwise from `.coin/manifest.local.yaml` relative to the project root. The file MUST exist at resolve time.

#### Scenario: Default manifest file path

- **WHEN** product config sets `coin.resolve: file` and omits `coin.manifestFile`
- **THEN** coin-lib MUST read `.coin/manifest.local.yaml`

#### Scenario: Custom manifest file path

- **WHEN** product config sets `coin.resolve: file` and `coin.manifestFile: .coin/fixtures/go-app.yaml`
- **THEN** coin-lib MUST read that path
- **AND** MUST fail if the file is missing

### Requirement: Soft warning for file resolve

When resolving with `coin.resolve: file`, coin-lib SHALL emit a visible soft warning in the Jenkins log that local manifest resolve is intended for platform/sample debugging. coin-lib MUST NOT hard-fail solely because `resolve` is `file`.

#### Scenario: Soft warn on file resolve

- **WHEN** Jenkins resolves with `coin.resolve: file`
- **THEN** the build log MUST contain a warning about local/file manifest resolve
- **AND** resolve MUST continue if the fixture file is valid

### Requirement: File resolve materializes runtime manifest

After loading the fixture, coin-lib SHALL materialize `.coin/manifest.json` for coin-executor the same way as after remote resolve. Fixture content MUST be a resolved manifest shape consumable by coin-executor (not a partial GP authoring body alone).

#### Scenario: Fixture becomes runtime manifest

- **WHEN** file resolve succeeds
- **THEN** workspace MUST contain `.coin/manifest.json` with the loaded manifest
- **AND** subsequent `coin-executor` invocations MUST use that file via `--manifest`
