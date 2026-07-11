## Why

GP release собирается из нескольких platform components (agent, executor, lib, gp-content, branching-model), но сегодня каждый тип публикуется своим путём: git + shell, разные форматы metadata, dual-write в PostgreSQL и Nexus. Enabling team не имеет единого playbook и вынуждена работать через git вместо control plane.

Принято решение **UI-first**: draft → canary на pilot GP → promote stable. SoT — coin-api + Nexus; git только optional export.

**Prerequisite для:** [gp-branching-model](../gp-branching-model/) (5-й composition slot).

## What Changes

- **Component Package Model** — единый Nexus layout, `content_ref` v2, generic materializers в resolve
- **Component lifecycle** — статусы `draft` | `canary` | `published`; product CI не видит draft
- **Component Studio** в coin-ui — primary path публикации platform components (не git/shell)
- **Canary promote pipeline** — единый wizard: component + catalog `latest_canary` → `latest` + health gate
- **Deprecate** git/Gitea platform publish path, `gp_artifact_bodies` dual-write, embedded seed как primary
- **BREAKING (future):** lib section в manifest; расширение composition slots (branching-model — отдельный change)

## Capabilities

### New Capabilities

- `component-platform`: lifecycle API, package model, resolve materializers, catalog/canary rules
- `component-studio`: coin-ui authoring — draft, validate, publish canary, promote stable

### Modified Capabilities

- _(пусто — baseline specs ещё не в `openspec/specs/`; после archive build-engine-model добавить deltas)_

## Impact

- **coin-api** — Admin API, manifest builder, schemas
- **coin-ui** — Component Studio, GP wizard, promote flow
- **Nexus** — unified package layout
- **PostgreSQL** — component_versions, catalog_policy, deprecate dual-write bodies
- **Enabling team workflow** — git/shell → UI only
- **Prerequisite для** [gp-branching-model](../gp-branching-model/) (5-й composition slot)

## Non-goals

- Fleet rollout 50/500/1500 repos (corp gate)
- Полная миграция gp-content в Studio до green field branching-model (GCP-3)
- Product repo branching (отдельный change gp-branching-model)
