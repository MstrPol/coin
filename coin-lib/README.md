# coin-lib

Jenkins Shared Library — **тонкий оркестратор** Coin CI.

## Ответственность

`coin-lib` выполняет ровно две задачи:

1. **Подготовка динамического агента** — читает `project.stack` и `runtime` из `.coin/config.yaml`, выбирает нужный K8s toolchain-образ из `resources/images.yaml`, запускает pod.
2. **Подготовка credentials** — биндит Jenkins Credentials перед вызовом `coin CLI`.

Вся остальная логика (версионирование, валидация, сборка, публикация, release notes) — в `coin-cli`.

## Структура

```
coin-lib/
├── vars/
│   └── coinPipeline.groovy      # единая точка входа
├── src/org/coin/ci/
│   ├── Config.groovy            # минимальное чтение config.yaml для оркестрации
│   ├── StackImages.groovy       # выбор образа агента из images.yaml
│   └── PodTemplate.groovy       # генерация K8s pod spec
└── resources/
    └── images.yaml              # каталог: stack → образ агента → версия coin CLI
```

## Использование в проекте

```groovy
@Library('coin-lib@1') _

coinPipeline()
```

Конфигурация — в `.coin/config.yaml`.

## Что НЕ добавлять в coin-lib

См. правило `.cursor/rules/coin-lib-scope.mdc`:
любая полезная логика идёт в `coin-cli`, не в Groovy.
