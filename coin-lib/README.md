# coin-lib

Jenkins Shared Library — **тонкий оркестратор** Coin CI.

## Ответственность

1. **Dynamic agent** — checkout `coin-platform` + project `.coin/config.yaml` → GP profile + `agents/catalog.yaml`.
2. **Credentials binding** перед `coin run publish`.

Вся логика (validate, version, test/build/publish) — в **coin-cli** + **coin-platform/golden-paths**.

## Разрешение образа агента

```
coin.template  →  platform/golden-paths/<tpl>/<ver>/profile.yaml  →  stack, runtime
jenkins.runtime / jenkins.agent.image (optional overrides)
agents/catalog.yaml  →  image ref
platform.yaml  →  jnlp image
```

## Структура

```
coin-lib/
  vars/coinPipeline.groovy
  src/org/coin/ci/
    Config.groovy
    StackImages.groovy
    PodTemplate.groovy
```

См. `.cursor/rules/coin-lib-scope.mdc`.
