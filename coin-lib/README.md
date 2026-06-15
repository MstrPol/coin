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
2. **GP** — resolved manifest (`runtime`, `executor`, `pipeline.stages`, …)
3. **project** — `.coin/config.yaml` из repo

В pod workspace materialize:

- `.coin/manifest.json` — для `coin-executor`
- `.coin/effective-config.yaml` — merged Jenkins glue config (debug)

`coinPodYaml` подставляет в [`coin-pod-template.yaml`](resources/coin-pod-template.yaml) образы и лимиты из merged config; defaults ресурсов — секция `pod` в `coin-lib-defaults.yaml`.

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

## Phase 1 (local pilot)

Lib загружается из Gitea `coin/coin-lib` с тегом **1.0.0** (для local test всегда один тег).

```bash
cd docker && make coin-lib
```

## Phase 2 (target)

Nexus immutable ZIP:

```
maven-releases/coin/lib/coin-lib/{version}/coin-lib-{version}.zip
```

```bash
cd docker && make coin-lib-http
```
