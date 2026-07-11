## Context

Agent registry (ADR `coin-ci-runtime`, spec `runtime-agent-registry`):

```
metadata.image  = nexus:8082/coin-docker/{profile}:{tag}
component_versions.version = {tag}   ← MUST match on promote (today)
```

CI: `VERSION` → build tag → POST draft с тем же version.  
UI catch-up: три поля, version часто расходится с tag.

## Goals / Non-Goals

**Goals:**

- Single SoT для manual path: **image ref** (tag = version).
- Server-side parse + validate на create.
- UI: Image + Digest only; preview derived version.

**Non-goals:**

- Decoupled platform version ≠ docker tag.
- Digest-only image refs without tag.
- Auto-bump / next-version picker для agent (semver из registry).

## Decisions

### D1. Parse rule (image tag → version)

Из `metadata.image` извлечь tag:

1. Отбросить digest suffix `@sha256:...` если есть.
2. Взять substring после последнего `/` (repo path).
3. Tag = substring после последнего `:` в этом сегменте.
4. Tag MUST be non-empty.

Примеры:

| Image ref | Version |
|-----------|---------|
| `nexus:8082/coin-docker/agent-30-06:1.2.0` | `1.2.0` |
| `nexus:8082/coin-docker/coin-agent:0.1.0-draft` | `0.1.0-draft` |

Reject: нет `:tag`, `tag` пустой, `latest`, `sha256:...` как tag.

### D2. API — agent draft create

`POST /v1/admin/components/agent/{name}/versions/drafts`

| Поле body | Правило |
|-----------|---------|
| `metadata.image` | required |
| `metadata.digest` | optional at create; required at promote (unchanged) |
| `version` | **MUST NOT** be sent for agent manual register; if sent — **422** «use image tag» (avoid silent mismatch). CI `publish-agent.sh` continues sending `version` + matching image — allowed when tag matches. |

Server flow:

```
parse tag from image → version
assert image path contains /{profile}:
insert component_versions(version, metadata)
```

Validate on create (same as promote for tag/profile/digest format where applicable).

### D3. Profile ↔ repository name

Image repository segment (между последним `/` и `:`) MUST equal `components.name` (profile).

`agent-30-06` → `.../agent-30-06:1.2.0` ✅  
`.../coin-agent:1.2.0` для profile `agent-30-06` → 422.

### D4. UI — New draft (agent)

```
Image ref   [required]
Digest      [required]
Version     [read-only preview: "1.2.0" — from tag]
```

Submit: POST без `version` в body (или UI не отправляет version — API derives).

Live preview: parse on blur/change client-side (same rules); disable submit if parse fails.

### D5. UI — Edit metadata (draft)

`component_versions.version` immutable (PK).  
PATCH metadata: image tag MUST still equal existing version.  
Смена tag → пользователь создаёт **новый** draft (New draft), не edit.

### D6. CI path unchanged

`publish-agent.sh` передаёт `version` + `image` с тем же tag — coin-api проверяет совпадение; derive не ломает CI.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Нестандартные registry URL | Parse only last `/` segment; document examples |
| Клиенты шлют version без image | 422 с понятным message |
| Edit image с другим tag | PATCH validation tag == version |

## Migration Plan

1. coin-api derive + validation (backward: CI still works if version matches tag).
2. coin-ui form.
3. Docs.

No PG migration.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Reject `version` in body vs ignore | ✅ | Reject if present and ≠ parsed tag; CI may send matching version |
| Q2 | Semver-only tags | ✅ | v1: any non-empty tag except `latest`; semver recommended in docs |
