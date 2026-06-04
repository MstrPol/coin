# coin-lib

Jenkins Shared Library — **тонкий оркестратор** Coin CI.

## Ответственность

1. **Dynamic agent** — checkout `coin-platform` + project `.coin/config.yaml` → GP **profile bundle**.
2. **Bootstrap coin CLI** — Nexus Maven zip по `profile.coinCli.version`.
3. **Credentials binding** перед `coin run publish`.

Вся логика (validate, version, test/build/publish) — в **coin-cli** + **coin-platform/golden-paths**.

## Platform bundle (единый pin для продукта)

Продукт задаёт только `coin.template` + `coin.templateVersion`.  
Версии CLI и agent — в `golden-paths/<tpl>/<ver>/profile.yaml`:

```yaml
agent:
  stack: go
  runtime: { go: "1.22" }
  rev: 2
coinCli:
  version: "0.0.0-SNAPSHOT"
```

## Разрешение образа и CLI

```
coin.template + templateVersion  →  profile.yaml
  agent.stack/runtime/rev        →  agents/catalog.yaml  →  K8s image
  coinCli.version                →  Nexus Maven          →  .coin/bin/coin
platform.yaml                    →  jnlp, nexus, coinCli.min (пол)
jenkins.runtime / jenkins.agent.image  →  optional overrides проекта
```

## Структура

```
coin-lib/
  vars/coinPipeline.groovy
  src/org/coin/ci/
    Config.groovy
    ProfileLoader.groovy
    StackImages.groovy
    CoinCli.groovy
    PodTemplate.groovy
    Platform.groovy
```

См. `.cursor/rules/coin-lib-scope.mdc`.
