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
  artifactId: my-service
```

---

## Секция `coin`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `coin.goldenPath` | **Да** | Имя GP: `go-app`, … |
| `coin.version` | **Да** | Semver **pin** GP (см. ниже) |
| `coin.resolve` | Нет | `remote` (default) или `file` — локальный fixture |
| `coin.manifestFile` | Нет | Путь к fixture при `resolve: file` (default `.coin/manifest.local.yaml`) |

Local resolve: [how-to/local-manifest-file-resolve.md](how-to/local-manifest-file-resolve.md).

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

В product config обычно задают только переопределения. Полный набор credential keys
собирается merge'ем с `coin-lib` defaults (см. ниже).

| Поле | Обязательно в product | Описание |
|------|----------------------|----------|
| `jenkins.credentials.docker` | **Да** (validate) | Jenkins Credential ID для OCI registry (pull/push) |

Agent image, executor, pipeline — **только в manifest**, не в config.

---

## Контракт Jenkins Credentials (platform)

Значения — **IDs записей в Jenkins Credentials Store**, не секреты.
Секреты живут только в Jenkins; в Git — только имена ID.

После merge (`lib` ← `project`) эффективный конфиг **должен** резолвить ключи контракта
(см. таблицу). Без соответствующих credentials в Jenkins пайплайн Coin неработоспособен
(для ключей, которые glue уже биндит). Зарезервированные ключи обязаны быть в контракте
заранее, даже если local pilot их ещё не использует.

| Ключ | Default ID (local) | Зачем | Кто биндит | Env / артефакт |
|------|--------------------|-------|------------|----------------|
| `apiToken` | `coin-api-token` | HTTP к coin-api (resolve, report) | coin-lib Resolve / Report | `COIN_API_TOKEN` |
| `nexus` | `nexus-admin` | Maven/raw Nexus (corp) | **ADR:** materialize → BuildKit secret `coin-nexus-*` | `/run/secrets/coin-nexus-user\|password` |
| `docker` | `nexus-docker` | OCI registry | build/publish | `COIN_REGISTRY_*` → `~/.docker/config.json` |
| `git` | `gitea-git` | fetch/push tags | Version / ReleaseNotes / Publish | `COIN_GIT_*` |
| `osc` | `osc-proxy` | Corp HTTP(S) proxy (OSC) | **ADR:** materialize → BuildKit secret `coin-osc-*` | `/run/secrets/coin-osc-user\|password` |

Proxy **host** (без пароля): build-arg `COIN_CORP_PROXY_URL` ← `coin.corpProxyUrl` (lib defaults).  
Nexus **deps URL** (без пароля): build-arg `COIN_NEXUS_URL` ← первый `manifest.destinations[].artifactRepositoryBase`.  
Auth: secret mounts `coin-osc-*` / `coin-nexus-*`. Детали: [ADR buildkit-secrets](adr/buildkit-secrets.md).

Defaults: [`coin-lib/resources/coin-lib-defaults.yaml`](../../coin-lib/resources/coin-lib-defaults.yaml).  
Создание ID в Jenkins: [jenkins-setup.md](jenkins-setup.md).

### Переопределение в проекте

```yaml
jenkins:
  credentials:
    docker: team-docker-registry   # свой ID вместо nexus-docker
    # git / apiToken / nexus / osc — опционально; иначе defaults из lib
```

Merge: `deepMerge` — project побеждает по ключу, остальные ключи остаются из lib.

### Кастомные ключи

| Где добавить | Что произойдёт |
|--------------|----------------|
| **Project** `.coin/config.yaml` → `jenkins.credentials.<custom>` | Попадёт в `.coin/effective-config.yaml` (merge). **Не** биндится автоматически. `coin-executor validate` **не** отвергает неизвестные ключи внутри `credentials` (**решение: no-op** / forward-compat). Фактически мёртвый конфиг, пока platform не научит glue читать ключ. |
| **GP / manifest** | **Нельзя.** Manifest — SoT runtime/pipeline/destinations, **не** Jenkins credential IDs. `coinLoadConfig` не переносит credentials из manifest. Класть секреты/ID в GP release запрещено контрактом. |

Новый credential для платформенного сценария = изменение **platform**:

1. ключ + default в `coin-lib-defaults.yaml`;
2. `withCredentials` / использование в `coinPipeline` (или узкий glue var);
3. запись в эту таблицу + JCasC/seed credentials на стенде;
4. если нужен внутри **Containerfile** — BuildKit secret mount ([ADR buildkit-secrets](adr/buildkit-secrets.md)).

Product-only «добавить credential и ожидать, что stages его увидят» — **не поддерживается**.

### Открытые вопросы / решения

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Неизвестные `jenkins.credentials.*` в validate | ✅ | **no-op** (forward-compat) |
| Q2 | Как `nexus` / `osc` в Containerfile? | ✅ | **BuildKit `RUN --mount=type=secret`** + Jenkins `withCredentials` → `.coin/secrets/` → executor `--secret`. См. [ADR buildkit-secrets](adr/buildkit-secrets.md). |

**Кратко Q2:** lib материализует `coin-nexus-*` / `coin-osc-*`; executor передаёт в `buildctl`; GP Containerfile монтирует по стабильным id. Local samples без mount. Реализация — follow-up checklist в ADR (пока только контракт).

---

## Секция `project`

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `project.name` | **Да** | Имя сервиса |
| `project.groupId` | **Да** | Домен команды |
| `project.artifactId` | **Да** | Artifact ID сервиса |

`project` не содержит destination fields: `repository`, `imageRepository`, `dockerRepository`, `mavenRepository`, `pypiRepository`.

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
| Deliverables | GP/Build Stack → `manifest.capabilities.deliverables` |
| Image registry prefix | `manifest.destinations.imageRegistryPrefix` |
| Artifact repository base URL | `manifest.destinations.artifactRepositoryBase` |
| Build cache on/off | `manifest.destinations.buildCacheEnabled` |

Product config также не содержит секции `deliverables`, build/publish commands, cache refs или registry/repository URLs.

---

## Jenkins glue config layers

`coin-lib` собирает runtime-конфиг для Jenkins из трёх слоёв (поздний побеждает):

| Слой | Источник | Примеры |
|------|----------|---------|
| lib | `coin-lib/resources/coin-lib-defaults.yaml` + env | `coin.apiUrl`, полный набор `jenkins.credentials.*` |
| GP | resolved `manifest.json` | `runtime.image`, `pipeline.tasks`, `build.engine`, `destinations` — **без** credentials |
| project | `.coin/config.yaml` | `coin.goldenPath`, `project.*`, override `jenkins.credentials.*` |

В workspace pod пишутся runtime artifacts (в `.gitignore`):

- `.coin/manifest.json` — для `coin-executor`
- `.coin/effective-config.yaml` — merged Jenkins glue (debug)

`coin-executor` и GP scripts читают **project** `.coin/config.yaml`, не effective config.

Resolved manifest **не** содержит Jenkins credential IDs. Контракт credentials — [выше](#контракт-jenkins-credentials-platform).

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
