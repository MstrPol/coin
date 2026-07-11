## 1. coin-api

- [x] 1.1 `DeleteComponentVersionDraft` в store (draft-only, cascade, audit `delete_component_draft`)
- [x] 1.2 Admin service + `DELETE /v1/admin/components/{type}/{name}/versions/{version}` handler (204/404/409)
- [x] 1.3 OpenAPI: document delete endpoint
- [x] 1.4 Unit tests: delete draft, reject published, not found

## 2. coin-ui (runtime agent)

- [x] 2.1 `api.deleteComponentVersionDraft(type, name, version)`
- [x] 2.2 `PlatformReleasesTab` — Delete draft per row для agent
- [x] 2.3 `PlatformComponentReleaseDetail` — Delete draft button + confirm + redirect
- [x] 2.4 Publisher gate (`can("publisher")`)

## 3. E2E + docs

- [x] 3.1 `e2e-platform-component-hub.sh` — create draft → delete → verify gone
- [x] 3.2 Docs: `agent-build-model.md` / `coin-ui-user-guide.md` — cleanup orphan drafts

## 4. OpenSpec

- [x] 4.1 `openspec validate runtime-agent-draft-delete --strict`
- [x] 4.2 Archive + baseline sync (после apply)
