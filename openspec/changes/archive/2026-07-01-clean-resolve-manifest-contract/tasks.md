## 1. Contract and Schema

- [x] 1.1 Remove top-level `credentials` from `coin-api/manifest.schema.json`
- [x] 1.2 Update manifest schema tests/fixtures to reject resolved manifests containing Jenkins credential IDs
- [x] 1.3 Keep `manifestVersion` at `1` and document this as a local pilot hard cut

## 2. coin-api Resolve

- [x] 2.1 Remove hardcoded `credentials: { docker: "nexus-docker" }` from `coin-api/internal/manifest/builder.go`
- [x] 2.2 Add/update builder tests proving resolve output contains only GP identity, `runtime`, `build`, `pipeline`, `validateSchema`, `capabilities`, `branching`, and integrity metadata
- [x] 2.3 Verify resolve with product context still resolves project-specific `cacheRef` without adding Jenkins/site fields

## 3. coin-lib Boundary

- [x] 3.1 Remove `manifest.credentials` merge from `coin-lib/vars/coinLoadConfig.groovy`
- [x] 3.2 Ensure `coinPipeline` binds Docker credentials from product config or `coin-lib` defaults when manifest has no `credentials`
- [x] 3.3 Update or add coin-lib tests/fixtures for manifest without credentials

## 4. Documentation and Samples

- [x] 4.1 Update docs that show `manifest.credentials` as part of resolved manifest
- [x] 4.2 Keep product/starters `.coin/config.yaml` examples as the Jenkins credential source
- [x] 4.3 Update troubleshoot/onboarding docs to distinguish product/Jenkins credentials from GP manifest fields

## 5. Verification

- [x] 5.1 Run relevant `coin-api` tests for manifest builder/schema
- [x] 5.2 Run relevant `coin-lib` validation/tests or dry-run flow
- [x] 5.3 Re-resolve `gp-01-07@1.0.0` with `project=demo-go-app` and confirm the manifest has no `credentials` key
- [x] 5.4 Confirm `samples/demo-go-app` CI still gets Docker credentials through product config/defaults
