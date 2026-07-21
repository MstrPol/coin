# Build engines и runtime agent

Operational runbook. **Каноническая модель:** ADR [coin-ci-runtime](adr/coin-ci-runtime.md). Решение о `build.engine`: [build-engine-contract](adr/build-engine-contract.md).

Продукт **не** выбирает engine — он зафиксирован в embedded pipeline GP release и приходит в manifest как `build.engine` / pipeline stages.

---

## Компоненты

| Слой | Роль |
|------|------|
| **GP release (embedded pipeline)** | SoT: engines, stages/steps, managed Containerfiles, schema |
| **coin-api** | Resolve → manifest; bootstrap seed |
| **coin-agent** | Единый Jenkins inbound-agent image: `coin-executor`, `podman`, buildkit binaries |
| **coin-executor** | `validate` / `run --stage` / `publish` |
| **coin-lib** | Jenkins glue: resolve, pod, credentials, bootstrap podman, вызов executor |

**Не используется:** language-specific stack agents, GP `scripts/*.sh` в runtime, dual-container pod, bootstrap `curl coin-executor`, папка `coin-gp-content/`.

---

## Два build engine (local pilot)

| Engine | Golden path (sample) | Сборка образа | Containerfile |
|--------|----------------------|---------------|---------------|
| `buildkit` | `go-app` | BuildKit targets через **podman build** | Managed из GP pipeline / manifest |
| `dockerfile` (BYO) | `go-app-docker` | **podman build** по project Dockerfile | Dockerfile в репозитории продукта |

Эталон seed (bootstrap only):

- [`coin-api/internal/gpcontent/seed/pipelines/go-app.yaml`](../../coin-api/internal/gpcontent/seed/pipelines/go-app.yaml)
- [`coin-api/internal/gpcontent/seed/pipelines/go-app-docker.yaml`](../../coin-api/internal/gpcontent/seed/pipelines/go-app-docker.yaml)

Live authoring: GP release Pipeline section в Platform UI ([gp-embedded-pipeline](adr/gp-embedded-pipeline.md)).

E2E на стенде:

```bash
cd docker
make seed-jenkins-lib          # lib + branching + GP profiles (embedded pipeline)
make publish-agent             # coin-agent → Nexus (arm64: GOARCH=arm64)
make samples                   # demo-go-app, demo-go-app-docker
make e2e-build-engines         # обе job → SUCCESS
```

---

## Поток CI

```mermaid
flowchart TB
  subgraph jenkins["Jenkins controller"]
    R["coin-lib: Resolve manifest"]
  end

  subgraph pod["K8s pod: один container jnlp = coin-agent"]
    B["Bootstrap: podman system service"]
    V["coin-executor run --stage validate"]
    T["coin-executor run --stage test"]
    BI["coin-executor run --stage build"]
    P["coin-executor run --stage publish"]
    REP["coin-executor report"]
  end

  R --> pod
  B --> V --> T --> BI --> P --> REP
  BI --> REG["Nexus Docker registry"]
```

### Стадии (typed, без script URL)

| Stage | Поведение |
|-------|-----------|
| `validate` | JSON schema `.coin/config.yaml` + manifest capabilities |
| `test` | buildkit: target `test` в Containerfile; BYO dockerfile: `testTarget` в product Dockerfile |
| `build` | Сборка image → `.coin/outputs.json` |
| `publish` | Push image; skip — `params.publish=false` (coin-lib); deny — branching policy |

Orchestration — только `coin-lib` (`coinPipeline()`), не Groovy из Nexus.

---

## Manifest: секция `build`

Схема: [`coin-api/manifest.schema.json`](../coin-api/manifest.schema.json).

Пример (`go-app`, buildkit):

```json
{
  "destinations": {
    "imageRegistryPrefix": "localhost:8082/coin-docker",
    "buildCacheEnabled": true,
    "artifactRepositoryBase": "http://nexus:8081/repository/maven-releases"
  },
  "build": {
    "engine": "buildkit",
    "buildkit": {
      "dockerfile": ".coin/Containerfile",
      "targets": {
        "validate": "validate",
        "test": "test",
        "image": "runtime",
        "artifact": "artifact"
      },
      "containerfile": {
        "url": "http://nexus:8081/repository/maven-releases/coin/gp/content/go-app/1.0.2/...",
        "sha256": "sha256:..."
      }
    }
  },
  "pipeline": {
    "stages": [
      { "id": "validate", "name": "Validate" },
      { "id": "test", "name": "Test" },
      { "id": "build", "name": "Build" },
      { "id": "publish", "name": "Publish" }
    ]
  }
}
```

Publish gate: Jenkins `params.publish` → `COIN_PUBLISH_REQUEST`; ветка — `manifest.branching`. См. [adr/gp-branching-model.md](adr/gp-branching-model.md).

`buildpack` и `dockerfile` — поля `build.buildpack` / `build.dockerfile` (см. schema).

Managed Containerfile materialize в workspace: `.coin/Containerfile` (путь из `build.*.dockerfile`).

---

## coin-agent image

Кратко — см. [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md) §2. Собирается из [`coin-executor/Dockerfile.agent`](../coin-executor/Dockerfile.agent), публикуется скриптом `coin-executor/scripts/publish-agent.sh`.

| Компонент | Назначение |
|-----------|------------|
| `jenkins/inbound-agent` | JNLP remoting |
| `coin-executor` | Baked binary (не curl в bootstrap) |
| `podman` | Container builds + socket для `pack` |
| `pack` | Buildpack engine |
| `buildkitd` / `buildctl` | В образе для fallback; **не** стартуют в bootstrap на local pilot |
| `paketo-builder.tar` | Baked builder для buildpack (load в bootstrap) |

Registry (runtime): `nexus:8082/coin-docker/coin-agent:{semver}`.

**Publish flow (draft → published):**

1. `publish-agent.sh` builds and pushes image to Nexus Docker.
2. `POST /v1/admin/components/agent/coin-agent/versions/drafts` — register draft with metadata `{image, digest}` (без `goarch`).
3. **Promote только вручную** — Platform UI или `POST .../versions/{version}/promote` (publisher). CI **не** auto-promote.
4. Platform UI (`/platform/runtime/coin-agent`) — CI path: read-only metadata + Publish; manual catch-up: **Image ref + Digest** (version = тег образа, preview read-only).

Promote gate: coin-api отклоняет promote без `metadata.image` и `metadata.digest` (HTTP 422); тег в image MUST совпадать с `component_versions.version` (version derive из image при create).

**Cleanup:** orphan drafts (тестовые профили, неудачный CI register) — удалить в Platform UI: `/platform/runtime/{profile}/releases` → Delete, или `DELETE /v1/admin/components/agent/{profile}/versions/{version}` (только `status=draft`).

Composition slot: `agent` → `coin-agent@version` (не stack-specific images).

---

## Pod template (coin-lib)

[`coin-lib/resources/coin-pod-template.yaml`](../coin-lib/resources/coin-pod-template.yaml):

- один container `jnlp` = `manifest.runtime.image`
- `securityContext.privileged: true`, `procMount: Unmasked` (nested containers)
- `emptyDir` 12Gi → `/var/lib/containers/storage` (podman graph)
- env `COIN_BUILD_ENGINE` из manifest (`buildkit` | `buildpack` | `dockerfile`)

---

## Bootstrap (coinPipeline)

1. `podman system service` → `unix:///var/run/docker.sock`
2. **buildpack only:** `podman load` из `/usr/share/coin/paketo-builder.tar` (если builder ещё не в store)
3. `coin-executor version`

`buildkitd` **не** запускается на local pilot arm64 — см. ниже.

---

## Реализация engine в coin-executor

| Engine | Implementation |
|--------|----------------|
| `buildkit` | `RunTarget()` → **podman build** если socket есть, иначе `buildctl` |
| `dockerfile` | то же (другие targets из manifest) |
| `buildpack` | `pack build` с `--docker-host inherit`, `--network host`, pinned `BP_GO_VERSION` |

Buildpack на arm64 pilot: builder jammy-base amd64, работает через baked tar + host network.

### Local pilot arm64: почему podman, а не buildctl

В k3s на Apple Silicon `buildctl` + `RUN` в Dockerfile даёт `exec /bin/sh: invalid argument` (nested runc). **Обходной путь pilot:** container builds через podman в том же privileged pod. Контракт `build.engine: buildkit` сохранён; implementation — podman на этом стенде.

На amd64 corp (roadmap) — нативный buildkitd + buildctl без podman-fallback.

---

## Registry cache

Registry cache ref вычисляет `coin-executor` из `manifest.destinations.imageRegistryPrefix`, `project.groupId`, `project.artifactId`, `project.name` и suffix `-cache`.

Пример: `localhost:8082/coin-docker/com.example.team/demo-go-app/demo-go-app-cache`.

---

## Операционные заметки (local pilot)

| Симптом | Причина | Действие |
|---------|---------|----------|
| Pod `TerminationByKubelet` ephemeral-storage | Диск k3s полон (agent image ~3GiB + podman load) | `make e2e-build-engines` (с prune) или `bash docker/scripts/prune-k3s-disk.sh --all` |
| `short-name golang:… did not resolve` | podman registries.conf | В agent image: `unqualified-search-registries = ["docker.io"]` |
| manifest sha256 mismatch | Nexus immutable + новый GP release content | **Новый GP release** (не UPDATE blob) |
| Старый bootstrap (buildkitd) в логе | Jenkins cache Shared Library | `make coin-lib` (очищает `caches/git-*`) |
| `component_version` seed берёт 1.0.0 | API order ≠ semver max | Исправлено в `seed-jenkins-lib-stack.sh` |

---

## Superseded (не документировать / не реализовывать)

- `coin-jenkins-agents/`, `agents-build`, stack catalog `images.yaml`
- Native compile в language agent + `pack-image.sh`
- `coin-golden-paths/` monolith tree с `scripts/test.sh`
- Host `docker.sock` mount
- `manifest.pipeline.stages[].script.url`
- `manifest.jnlp` / отдельный stack container

---

## См. также

- [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md) — canonical CI runtime ADR
- [golden-paths.md](golden-paths.md) — composition, samples
- [jenkins-setup.md](jenkins-setup.md) — credentials, platform jobs
- [config.md](config.md) — что в проекте vs manifest
- [how-to/troubleshoot-ci.md](how-to/troubleshoot-ci.md) — CI errors
- [docker/README.md](../docker/README.md) — make targets, стенд
