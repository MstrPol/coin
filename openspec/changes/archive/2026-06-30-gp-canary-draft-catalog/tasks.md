## 1. coin-api

- [x] 1.1 `ValidateCatalogPolicyUpdate`: `latestCanary` — draft или published GP release
- [x] 1.2 `resolve`: `AllowDraftGP` при `channel=canary`

## 2. coin-ui

- [x] 2.1 `GpPolicyTab`: canary picker — draft + published GP versions
- [x] 2.2 Warning при выборе GP draft на canary line

## 3. Validation

- [x] 3.1 `go test ./internal/store/...`
- [x] 3.2 Manual: Policy → Latest canary → GP draft → Save → Canary preview resolve
