# coin-lib

Jenkins Shared Library — **тонкий оркестратор** Coin CI.

## Ответственность

`coin-lib` выполняет ровно две задачи:

1. **Подготовка dynamic agent** — по `coin.template` → stack, `jenkins.runtime` → version, lookup в `resources/images.yaml`, K8s pod (jnlp + stack).
2. **Credentials** — bind Jenkins Credentials перед `coin run publish`.

Вся логика (validate, version, test/build/publish, Dockerfile render) — в **coin-cli** + **coin-golden-paths**.

## Структура

```
coin-lib/
├── vars/
│   └── coinPipeline.groovy      # единая точка входа
├── src/org/coin/ci/
│   ├── Config.groovy            # чтение .coin/config.yaml
│   ├── StackImages.groovy       # template → stack → agent image
│   └── PodTemplate.groovy       # K8s pod spec
└── resources/
    └── images.yaml              # stacks, templates, jnlp, coinCli
```

## Разрешение agent image

```
coin.template  →  images.yaml templates  →  stack
profile defaults (GP) + jenkins.runtime  →  version key
images.yaml stacks[stack][version]       →  image ref (+ optional digest)
```

Подробнее — [docs/agent-build-model.md](../docs/agent-build-model.md).

## Использование в сервисе

```groovy
@Library('coin-lib') _

coinPipeline()
```

Конфигурация — `.coin/config.yaml` в репозитории **сервиса**.

## Что НЕ добавлять в coin-lib

См. `.cursor/rules/coin-lib-scope.mdc` — логика только в `coin-cli`.

## Связанные документы

- [docs/jenkins-setup.md](../docs/jenkins-setup.md)
- [docs/architecture.md](../docs/architecture.md)
