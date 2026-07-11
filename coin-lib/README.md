# coin-lib

Jenkins Shared Library — только glue: resolve manifest, K8s pod, credentials, вызов `coin-executor` stages.

Бизнес-логика сборки (validate, build, publish, versioning) — в `coin-executor`.

## Product Jenkinsfile

```groovy
@Library('coin-lib@1.0.0') _
coinPipeline()
```

Product repo не содержит platform API URL, `/version` lookup или credentials для bootstrap.
`coin-lib@1.0.0` resolve'ит manifest и запускает stages из `manifest.pipeline.stages`.

### Параметры сборки

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `publish` | boolean | `false` | Выполнить stage publish из manifest |

Обычные сборки (main, PR) идут без publish. Для публикации артефактов включите параметр вручную (**Build with Parameters**) или настройте job/trigger.

## Structure

```
coin-lib/
  resources/
    coin-lib-defaults.yaml
    coin-pod-template.yaml
  vars/
    coinLog.groovy           # structured console logging
    coinPipeline.groovy      # entrypoint
    coinPodYaml.groovy       # render pod template from config
    coinLoadConfig.groovy    # layered config merge
    coinReadProjectConfig.groovy
    coinResolveManifest.groovy
    coinMaterializeDotCoin.groovy
    coinApplyEnv.groovy
    coinRunStage.groovy
  scripts/publish-lib.sh
  Jenkinsfile
```

`src/` не используется — global vars + `libraryResource`, без `load` из product workspace.

## Layered config

Merge в `coinLoadConfig` (приоритет слева направо, поздний слой побеждает):

1. **lib** — `resources/coin-lib-defaults.yaml` + env overrides (`COIN_API_URL`, …)
2. **GP** — resolved manifest (`runtime`, `build.engine`, `pipeline.stages`, …)
3. **project** — `.coin/config.yaml` из repo (`jenkins.credentials.*` включительно)

В pod workspace materialize:

- `.coin/manifest.json` — для `coin-executor`
- `.coin/effective-config.yaml` — merged Jenkins glue config (debug)

`coinPodYaml` подставляет в [`coin-pod-template.yaml`](resources/coin-pod-template.yaml) образ `runtime.image`, env `COIN_BUILD_ENGINE`, privileged + `procMount: Unmasked`, emptyDir 12Gi для podman storage.

Resolved manifest не содержит Jenkins credential IDs. `coinPipeline` выбирает Docker credential из product config (`jenkins.credentials.docker`) или defaults `coin-lib`.

## Bootstrap (в pod)

`coinPipeline` на agent node:

1. `podman system service` → `unix:///var/run/docker.sock`
2. `coin-executor version`

`buildkitd` **не** стартует на local pilot arm64 — builds через podman. См. [docs/agent-build-model.md](../docs/agent-build-model.md).

## Console logging

Все сообщения coin-lib идут через `coinLog` — единый формат с emoji и визуальными разделителями стейджей:

```
════════════════════════════════════════════════════
🔍  Coin │ Resolve manifest
────────────────────────────────────────────────────
   ✅  Project config loaded via readTrusted (no checkout)
   ▶️  Resolve manifest: http://coin-api:8090/v1/...
   ✅  coin-api resolve HTTP 200
   🎯  Resolved GP: go-app@1.0.4
────────────────────────────────────────────────────
```

API: `coinLog.section(emoji, title)`, `coinLog.line`, `coinLog.kv`, `coinLog.ok`, `coinLog.warn`, `coinLog.skip`, `coinLog.step`, `coinLog.sectionEnd()`.

## Phase 1 (local pilot) — deprecated

**Deprecated:** lib загружается из Gitea `coin/coin-lib` с тегом **1.0.0** (Modern SCM retriever).

```bash
cd docker && make coin-lib   # bootstrap only — не primary path
```

## Primary path (Nexus HTTP)

Immutable ZIP в Nexus + Jenkins HTTP Shared Library retriever. **coin-api registry не используется** — см. [docs/adr/jenkins-lib-outside-platform.md](../docs/adr/jenkins-lib-outside-platform.md).

```
maven-releases/coin/lib/coin-lib/{version}/coin-lib-{version}.zip
```

```bash
# publish ZIP to Nexus only
./scripts/publish-lib.sh 1.0.0

# local stack: seed + switch Jenkins retriever
cd docker && make seed-jenkins-lib   # includes coin-lib-http

# или только переключить retriever после publish
cd docker && make coin-lib-http
```

Версия lib для product jobs — в Jenkins CASC / `@Library('coin-lib@…')`, не в coin-api manifest.
