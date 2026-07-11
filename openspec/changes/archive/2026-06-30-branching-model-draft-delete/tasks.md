## 1. coin-ui — Releases tab

- [x] 1.1 Добавить `supportsDraftDelete(compType)` (`agent` | `branching-model`) в shared helper или inline
- [x] 1.2 `PlatformReleasesTab`: показывать Delete для branching-model drafts (убрать `isAgent`-only guard)
- [x] 1.3 Confirm dialog + reload list после успешного delete

## 2. coin-ui — Editor lifecycle

- [x] 2.1 `PlatformComponentEditor` / `LifecyclePanel`: «Delete draft» для `branching-model` drafts (publisher+)
- [x] 2.2 После delete — navigate на `{hub}/releases` (`familyHubPath`)
- [x] 2.3 Optional: wire delete в `PlatformComponentReleaseDetail` если embedded editor не покрывает все пути — embedded `PlatformComponentEditor` покрывает branching-model draft release detail

## 3. E2E + docs

- [x] 3.1 `e2e-platform-component-hub.sh`: create bml draft → DELETE → verify 404
- [x] 3.2 Docs: `docs/how-to/branching-models.md` или `docs/coin-ui-user-guide.md` — cleanup orphan drafts

## 4. OpenSpec

- [x] 4.1 `openspec validate branching-model-draft-delete --strict`
- [x] 4.2 Archive + baseline sync (после apply)
