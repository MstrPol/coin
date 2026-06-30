## 1. API contract

- [x] 1.1 `validateAgentMetadata` — image + digest format, tag↔version invariant
- [x] 1.2 `PromoteComponentToPublished` для `agent` — reject без image/digest (422)
- [x] 1.3 OpenAPI: agent metadata без `goarch`; promote errors documented
- [x] 1.4 `executorPinForAgentStack` — same-V для любого profile; убрать hardcoded switch
- [x] 1.5 Тесты lifecycle + promote gate + multi-profile derive

## 2. CI / scripts

- [x] 2.1 `publish-agent.sh` — убрать auto-promote; draft register only
- [x] 2.2 Обновить вывод скрипта («promote in Platform UI»)
- [x] 2.3 `e2e-platform-component-hub.sh` — promote явным шагом
- [x] 2.4 `e2e-bootstrap.sh` / seed — без ожидания auto-published agent (если применимо)

## 3. coin-ui

- [x] 3.1 `PlatformNewDraftPage` — agent: version + image + digest required; без GOARCH
- [x] 3.2 `PlatformAgentMetadataEditorPage` — только image + digest; без GOARCH
- [x] 3.3 `PlatformComponentReleaseDetail` — убрать GOARCH; Promote CTA для draft
- [x] 3.4 Copy/hints: CI path vs manual catch-up

## 4. Docs + ADR

- [x] 4.1 Amend `docs/adr/coin-ci-runtime.md` — registry model, manual promote
- [x] 4.2 Update `docs/agent-build-model.md` — publish flow, drop goarch metadata
- [x] 4.3 Проверить docs на хвосты `goarch` в platform context / `/platform/components`

## 5. OpenSpec

- [x] 5.1 `openspec validate runtime-agent-registry --strict`
- [x] 5.2 Archive + baseline sync (после apply)
