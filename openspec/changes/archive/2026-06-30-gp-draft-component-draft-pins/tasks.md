## 1. Shared version helper

- [x] 1.1 Вынести `versionLabels` / `publishedVersions` в `coin-ui/src/lib/gpCompositionVersions.ts`
- [x] 1.2 Перевести `useGpCompositionEditor` на shared helper

## 2. coin-ui — PublishWizard

- [x] 2.1 Version pickers: gc/bm — draft+published; agent — published only
- [x] 2.2 Добавить `versionStatuses` в wizard; передать в `GpCompositionForm`
- [x] 2.3 Empty states: различать «нет published agent» vs «нет версий на hub»
- [x] 2.4 `draftBlocked`: блокировать только при отсутствии версий / published agent, не при draft gc/bm pins
- [x] 2.5 Warning panel draft pins (reuse `GpCompositionForm` или inline)

## 3. Verify promote gate (no regression)

- [x] 3.1 Confirm `GpReleaseDetail` promote disabled при `draftPinCount > 0` + 409 surfacing
- [x] 3.2 Manual smoke: create GP draft с draft bs/bm → promote blocked → publish pins → promote OK

## 4. E2E + docs

- [x] 4.1 E2E: GP draft с draft gc/bm pin + promote 409 (extend existing script или новый)
- [x] 4.2 `docs/coin-ui-user-guide.md` — workflow GP draft на component drafts

## 5. OpenSpec

- [x] 5.1 `openspec validate gp-draft-component-draft-pins --strict`
- [x] 5.2 Archive + baseline sync (после apply)
