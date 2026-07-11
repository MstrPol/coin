## 1. Canonical ADR

- [x] 1.1 Создать `docs/adr/coin-ci-runtime.md`: coin-agent состав, bootstrap, 3 engines, pilot vs corp, publish layers, three-pin composition, superseded list
- [x] 1.2 Amend `docs/adr/build-engine-contract.md`: banner + ссылка на `coin-ci-runtime`; контекст пометить как pre-hard-cut
- [x] 1.3 Обновить `docs/adr/README.md`: добавить `coin-ci-runtime`, `gp-branching-model`; проверить статусы superseded

## 2. Superseded ADR banners

- [x] 2.1 Banner в `docs/adr/gp-composition-four-components.md` (superseded → three-pin + coin-ci-runtime)
- [x] 2.2 Banner в `docs/adr/gp-pipeline-bundle-layer.md` (superseded → coin-lib + jenkins-lib ADR)

## 3. Top-level docs sync

- [x] 3.1 `docs/architecture.md`: three-pin composition, derived executor, ссылка на coin-ci-runtime; убрать 4-slot table
- [x] 3.2 `docs/control-plane.md`: manifest example без `when: tag`; branching + params.publish; dedupe engine tables → link ADR
- [x] 3.3 `docs/agent-build-model.md`: publish gate без `when: tag`; ссылка на coin-ci-runtime; убрать дубли с ADR где возможно

## 4. Legacy grep sweep

- [x] 4.1 Grep `docs/` на: `pipeline-bundle`, `manifest.jnlp`, `coin-jenkins-agents`, `when: tag`, `four component`, standalone `executor` slot
- [x] 4.2 Исправить high-traffic hits: `docs/README.md`, `docs/how-to/troubleshoot-ci.md`, `docs/golden-paths.md` (точечно)
- [x] 4.3 Проверить cross-links из `.cursor/rules/coin-project-gates.mdc` — OK: pipeline-bundle помечен superseded, согласуется с coin-ci-runtime

## 5. OpenSpec closure

- [x] 5.1 `openspec validate sync-runtime-docs-adr --strict`
- [x] 5.2 Archive change → baseline `openspec/specs/runtime-documentation/spec.md`
- [x] 5.3 Обновить Purpose в `openspec/specs/build-engine/spec.md` и `gp-composition-two-slot/spec.md` при archive

## 6. Gate (manual)

- [x] 6.1 Peer review: путь architecture → coin-ci-runtime → agent-build-model проверен (three-pin, publish gate, no 4-slot)
- [x] 6.2 gp-content reference YAML (`when: tag`, `controls`) — cleanup в change `gp-content-schema-v2` (зафиксировано в proposal/design non-goals)
