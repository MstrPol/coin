## MODIFIED Requirements

### Requirement: Pilot vs corp build implementation

`docs/adr/coin-ci-runtime.md` SHALL document two product engines (`buildkit`, BYO `dockerfile`) and that buildpack is superseded. On local pilot arm64 both engines use podman build when podman socket is available.

#### Scenario: Pilot troubleshooting

- **WHEN** a reader debugs arm64 pilot build failures
- **THEN** `docs/agent-build-model.md` MUST explain podman-first implementation for both engines
- **AND** MUST NOT list buildpack bootstrap steps
