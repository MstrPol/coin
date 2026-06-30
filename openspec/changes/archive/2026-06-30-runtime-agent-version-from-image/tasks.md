## 1. coin-api

- [x] 1.1 `parseAgentImageRef(image, profileName) → (version, error)` в store или internal package
- [x] 1.2 Agent draft create: derive version из image; reject mismatch если `version` в body ≠ tag
- [x] 1.3 Validate profile ↔ repo segment + tag rules на create (не только promote)
- [x] 1.4 PATCH agent metadata: image tag MUST equal existing version
- [x] 1.5 OpenAPI: agent draft — version derived; document 422 cases
- [x] 1.6 Unit tests parse + create validation

## 2. coin-ui

- [x] 2.1 `PlatformNewDraftPage` — убрать Version input; preview из image tag
- [x] 2.2 Client-side parse preview (shared helper с API rules)
- [x] 2.3 POST draft без `version` (или только derived server-side)
- [x] 2.4 `PlatformAgentMetadataEditorPage` — validate tag == version on save

## 3. Docs

- [x] 3.1 `agent-build-model.md` — manual catch-up без отдельного Version
- [x] 3.2 `coin-ui-user-guide.md` — форма Image + Digest

## 4. OpenSpec

- [x] 4.1 `openspec validate runtime-agent-version-from-image --strict`
- [x] 4.2 Archive + baseline sync (после apply)
