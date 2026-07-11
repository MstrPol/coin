## REMOVED Requirements

### Requirement: GP content preview API

**Reason**: Preview superseded by GP release pipeline preview endpoint.

**Migration**: Use `POST /v1/admin/golden-paths/{name}/versions/{version}/pipeline/preview` per `gp-embedded-pipeline`.
