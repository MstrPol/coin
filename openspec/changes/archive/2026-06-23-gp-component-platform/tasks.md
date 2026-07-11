# Tasks: gp-component-platform

## GCP-0 — ADR + contract (BLOCKING)

- [x] GCP-0.1 ADR `gp-component-package-model`: UI-first lifecycle, package model, content_ref v2
- [x] GCP-0.2 Инвентаризация путей по component type + таблица deprecations в ADR
- [x] GCP-0.3 Закрыть Q1–Q4 с platform lead (Q5/Q6 решены)

## GCP-1 — API + Studio scaffold

- [x] GCP-1.1 coin-api: component draft CRUD, publish-to-canary, promote-to-stable
- [x] GCP-1.2 coin-api: resolve rules per channel (draft invisible to product CI)
- [x] GCP-1.3 coin-api: `component-package.schema.json` + content_ref v2
- [x] GCP-1.4 coin-api: Admin API upload package + register (backend Studio)
- [x] GCP-1.5 coin-ui UI-01: Component Studio scaffold (type-aware forms)
- [x] GCP-1.6 coin-ui UI-04: validate → Nexus → register flow
- [x] GCP-1.7 coin-ui UI-06: pilot projects picker + health gate before promote

## GCP-2 — Promote wizard

- [x] GCP-2.1 coin-ui UI-05: promote catalog latest_canary → latest + checklist
- [x] GCP-2.2 Единый promote flow: component + catalog + GP

## GCP-3 — gp-content migration

- [x] GCP-3.1 coin-ui UI-02: gp-content editor
- [x] GCP-3.2 coin-api: generic loader по composition slots (убрать switch per type)
- [x] GCP-3.3 Deprecate publish-content.sh как primary path

## GCP-4 — lib + cleanup

- [x] GCP-4.1 lib section в manifest + Nexus HTTP ZIP
- [x] GCP-4.2 Deprecate git/Gitea platform publish path

## GCP-5 — Fleet cleanup

- [x] GCP-5.1 Migration plan: gp_artifact_bodies dual-write, embedded seed
- [x] GCP-5.2 docs/control-plane.md + golden-paths: единая модель для enabling team

## Docs + handoff

- [x] docs/control-plane.md + golden-paths: единая модель для enabling team (см. GCP-5.2)
