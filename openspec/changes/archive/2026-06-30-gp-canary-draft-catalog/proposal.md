## Why

GP draft с draft component pins создаётся успешно, но назначить его на canary line (`catalog.latest_canary`) нельзя: UI показывает только published GP versions, API отклоняет draft/snapshot в `ValidateCatalogPolicyUpdate`. Это противоречит спеке `gp-composition-two-slot` (canary line MAY указывать на GP `draft`).

## What Changes

- **coin-api:** `latestCanary` в catalog policy — любой существующий GP release (`draft` или `published`), включая `-snapshot.N` semver.
- **coin-api:** canary channel resolve — `AllowDraftGP` при `channel=canary`, не только при snapshot exact pin.
- **coin-ui:** Policy tab — picker `Latest canary` включает draft GP releases с меткой `(draft)`; warning при выборе draft.

### Non-goals

- Менять stable line (`latest`, `minimum`) — только published.
- Auto-promote GP draft при назначении на canary.

## Capabilities

### Modified Capabilities

- `gp-composition-two-slot`: catalog policy validation для `latest_canary`
- `gp-publish-flows`: Policy tab canary picker UX

## Impact

| Область | Изменения |
|---------|-----------|
| coin-api | `catalog_admin.go`, `resolve/service.go` |
| coin-ui | `GpPolicyTab.tsx` |
