## Why

Manual catch-up форма agent draft (`/platform/runtime/{profile}/releases/new-draft`) просит **Version** и **Image ref** отдельно, хотя в registry-модели version **равен** docker tag. Пользователь вводит `0.1.0-draft` и image с тегом `1.2.0` → promote падает на 422. Дублирование полей не даёт гибкости (platform version ≠ tag не поддерживается) — только ошибки.

CI path (`publish-agent.sh`) уже использует один источник (VERSION → tag → `component_versions.version`). Нужна симметрия для Path B.

## What Changes

- **coin-api:** для `agent` draft register — **derive `version` из `metadata.image` tag**; поле `version` в body **не принимается** (или игнорируется с warning — см. design: reject if sent and mismatched).
- **coin-api:** validate tag↔profile name + parse rules на **create** (fail fast), не только promote.
- **coin-ui:** форма New draft agent — только **Image ref** + **Digest**; version — read-only preview из тега.
- **coin-ui:** metadata editor — image tag MUST match existing version (нельзя сменить tag на другой version без нового draft).
- Docs: manual catch-up flow без отдельного Version.

### Non-goals

- Изменение CI `publish-agent.sh` контракта (может слать version явно — MUST совпадать с tag).
- gp-content / branching-model draft forms (version-first остаётся).
- Поддержка image-only pin без semver tag (`latest`, digest-only ref).
- Build-stacks visual editor (отдельный backlog).

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `runtime-agent-registry`: version = image tag; API derive on agent draft create.
- `platform-runtime-catalog`: форма catch-up без поля Version.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | `parseAgentImageTag`, agent draft create validation |
| **coin-ui** | `PlatformNewDraftPage`, `PlatformAgentMetadataEditorPage` |
| **OpenAPI** | agent draft request shape |
| **docs** | `agent-build-model.md`, `coin-ui-user-guide.md` |
