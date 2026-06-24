# ADR: Coin Control Plane v2

**Статус:** accepted  
**Дата:** 2026-06-05  
**Контекст:** fleet 1500+ repos; текущая модель lib + cli + platform@main + profile pins не масштабируется.

## Решение

Ввести **Control Plane** из трёх слоёв:

| Слой | SoT | Роль |
|------|-----|------|
| Content | Git `coin-platform/content/` | scripts, Dockerfile, JSON Schema |
| Metadata | PostgreSQL | composition GP, semver, registry, audit |
| Runtime cache | Nexus `maven-releases` / `maven-snapshots` | immutable manifest blobs + mutable pointers |

Новые компоненты:

- **coin-api** — Resolve + Admin + registry
- **coin-executor** — validate, run, version, report (наследник coin-cli runtime)
- **coin-ui** — admin SPA (фаза 2)

**Hard cut:** coin-lib и coin-cli выводятся из эксплуатации с P0; dual path не поддерживается.

## Продуктовый контракт v2

```yaml
coin:
  goldenPath: go-app   # было: template
  version: "1.0.0"     # было: templateVersion: v1
```

Strict v2 only — executor не читает v1.

### Mapping v1 → v2

| v1 | v2 |
|----|-----|
| `coin.template: go-app` | `coin.goldenPath: go-app` |
| `coin.templateVersion: v1` | `coin.version: "1.0.0"` |

Semver GP release не равен каталогу `v1/` — это отдельная версия платформы (`go-app@1.0.0`).

## CI flow

1. Jenkins читает `.coin/config.yaml` (goldenPath + version)
2. Resolve `coin-api` → manifest JSON (primary)
3. Fallback: Nexus `manifest-{gp}-{ver}.json` если API down
4. Pod agent image из `manifest.runtime`
5. `coin-executor run --manifest ...`

## Отклонённые альтернативы

| Альтернатива | Почему нет |
|--------------|------------|
| GP-only in git | нет registry/blast-radius при 1500 repos |
| Nexus zip bundle (CPR) | теряется GitOps, маскирует drift |
| Параллельная поддержка lib/cli | dual maintenance |
| Manifest cache в git | CI должен работать без live API/DB; Nexus — CDN-like path |

## Последствия

- Все product repos мигрируют config до первого build на новом path
- P0 pilot: go-app@1.0.0, demo-go-app E2E
- HA coin-api и corp PostgreSQL — post-pilot (P1-05 deferred)
