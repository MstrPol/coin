# trunk-based

Модель ветвления для сервисных Golden Path (`go-app`, `go-app-bp`, `go-app-df`).

- **Trunk:** `main` — единственная долгоживущая ветка интеграции.
- **Ветки:** `feature/*`, `bugfix/*`, `release/*` (Jira ID в имени).
- **Версия:** RC-теги `v*-rc-*` на `release/*`; snapshot на feature при включённом qualifier.
- **Publish:** только при RC-теге (`publish.when: tag`).

См. [docs/branching.md](../../docs/branching.md) и [docs/adr/gp-branching-model.md](../../docs/adr/gp-branching-model.md).
