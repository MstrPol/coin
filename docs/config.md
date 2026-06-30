# `.coin/config.yaml` (v2)

Контракт между продуктовой командой и Control Plane.

Поведение сборки задаётся **manifest** (resolve по `goldenPath` + `version`), а не полями проекта.

Schema: [`coin-api/internal/gpcontent/seed/schema/config.v2.schema.json`](../coin-api/internal/gpcontent/seed/schema/config.v2.schema.json).

---

## Эталонный пример (go-app)

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"

jenkins:
  credentials:
    docker: nexus-docker

project:
  name: my-service
  groupId: com.example.team
  repository: Nexus_PROD
```

---

## Секция `coin`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `coin.goldenPath` | **Да** | Имя GP: `go-app`, … |
| `coin.version` | **Да** | Semver **pin** GP (см. ниже) |

### Pin-синтаксис (MVP-2)

| Pin | Смысл |
|-----|-------|
| `"=1.0.0"` | Exact — frozen, immutable |
| `"~1.0.0"` | Последний patch в линии 1.0.x |
| `"^1.0.0"` | Последний minor в линии 1.x |
| `"*"` | Latest stable из catalog |
| `"1.0.0"` | Alias для `=1.0.0` (backward compat) |
| `"1.0.0-snapshot.1"` | Explicit draft/snapshot (exact only) |

Resolve API: `GET /v1/golden-paths/{gp}/resolve?pin=~1.0.0`

Версия — **контракт semver платформы**, не каталог `v1/` из legacy.

---

## Секция `jenkins`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `jenkins.credentials.docker` | **Да** | Jenkins Credential ID для Docker registry |

Agent image, executor version, pipeline stages — **только в manifest**, не в config.

Credentials → env при publish: `COIN_REGISTRY_USER`, `COIN_REGISTRY_PASSWORD`.

---

## Секция `project`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `project.name` | **Да** | Имя сервиса |
| `project.groupId` | **Да** | Домен команды |
| `project.repository` | **Да** | Логическое имя репозитория Nexus |

---

## Миграция с v1

| v1 | v2 |
|----|-----|
| `coin.template: go-app` | `coin.goldenPath: go-app` |
| `coin.templateVersion: v1` | `coin.version: "1.0.0"` |

Подробно: [how-to/migrate-config-v1-to-v2.md](how-to/migrate-config-v1-to-v2.md).

---

## Что **не** задаётся в проекте

| Поле | Где живёт |
|------|-----------|
| Agent image | `manifest.runtime.image` |
| Build engine | `manifest.build.engine` (`buildkit` \| `buildpack` \| `dockerfile`) |
| Containerfile / targets | `manifest.build.buildkit` / `build.dockerfile` / `build.buildpack` |
| Pipeline stages | `manifest.pipeline.stages` (typed `id`, без script URLs) |
| coin-executor CLI | Baked в agent image (не отдельный platform component) |
| Config schema | `manifest.validateSchema` |

---

## Jenkins glue config layers

`coin-lib` собирает runtime-конфиг для Jenkins из трёх слоёв (поздний побеждает):

| Слой | Источник | Примеры |
|------|----------|---------|
| lib | `coin-lib/resources/coin-lib-defaults.yaml` + env | `coin.apiUrl`, credential IDs, registry prefix |
| GP | resolved `manifest.json` | `runtime.image`, `pipeline.stages` |
| project | `.coin/config.yaml` | `coin.goldenPath`, `project.*`, `jenkins.credentials.docker` |

В workspace pod пишутся runtime artifacts (в `.gitignore`):

- `.coin/manifest.json` — для `coin-executor`
- `.coin/effective-config.yaml` — merged Jenkins glue (debug)

`coin-executor` и GP scripts читают **project** `.coin/config.yaml`, не effective config.

---

## Verify

```bash
coin-executor validate --project .coin/config.yaml --manifest .coin/manifest.json
```

Manifest получить:

```bash
curl -fsS http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest \
  -o .coin/manifest.json
```
