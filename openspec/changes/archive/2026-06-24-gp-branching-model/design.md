# Design: GP Branching Model

**Prerequisite:** [gp-component-platform](../archive/2026-06-23-gp-component-platform/) — ✅ archived 2026-06-23.

## Composition (5 slots)

| Slot | Type | Manifest section |
|------|------|------------------|
| agent | agent | `runtime` |
| executor | executor | `executor` |
| lib | lib | `lib` |
| gp-content | gp-content | `build`, `pipeline`, … |
| **branching-model** | branching-model | **`branching`** |

Profile задаёт **имя** (`trunk-based`); release pin'ит **версию** (`1.0.0`).

## Authoring (UI-first)

Enabling team создаёт `branching-model` в Component Studio (UI-03): draft → canary на pilot GP → promote stable. Git — optional export.

Каталог моделей: `coin-branching-models/models/<name>/model.yaml` + README.md (human docs).

## Resolve flow

1. Read composition slot `branching-model`
2. Load `model.yaml` from Nexus package
3. Emit `manifest.branching` (full rules + source ref)
4. Executor reads manifest only — no runtime catalog lookup

## Executor (`internal/branching/`)

| Function | Purpose |
|----------|---------|
| `ValidateBranch` | branch name vs pattern/types |
| `ResolveVersion` | COIN_VERSION from git tags + rules |
| `ShouldPublish` | policy from branching + GIT_BRANCH/TAG_NAME |
| `Bump` | future `coin version bump` |

**Fix:** `run --stage X` must enforce stage policy (no bypass when single stage).

**Image tag:** use COIN_VERSION from ResolveVersion, not goldenPath.version pin.

## GP mapping (local pilot)

| GP | Model | Sample |
|----|-------|--------|
| go-app, go-app-bp, go-app-df | trunk-based@1.0.0 | demo-go-app* |
| go-lib / java-maven-app | semver-tag@1.0.0 | unit/demo |

## ADR

- Создать `docs/adr/gp-branching-model.md`
- Amend: gp-composition-four-components → superseded by 5-slot (в новом ADR)

## Open gate (platform lead)

Approve 5-й slot + component type; overrides v1 — none.

## Risks

- Breaking composition — все GP на 5 slots
- `params.publish=true` vs branching — manual override только platform/debug (зафиксировать в ADR)
