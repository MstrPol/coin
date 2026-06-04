# CI agent images

Toolchain-образы для Jenkins dynamic agents (K8s pod, контейнер `stack`).

Часть [coin-platform](../README.md). Сборка — job **`agents-build`** (`agents/Jenkinsfile`).

## catalog.yaml

Manifest agent images. Job **пишет** `rev`, `tag`, `digest` после каждой сборки.

Полный ref: `{registry.default}/{image}:{tag}` → `nexus:8082/coin-docker/ci-go:1.22-r1`.

## Связь с golden paths

```
golden-paths/<tpl>/vN/profile.yaml   agent.stack + agent.runtime
         │
         ▼
agents/catalog.yaml                  image, tag, digest
         │
         ▼
coin-lib StackImages (COIN_PLATFORM_DIR)  →  pod template
```

Проверка связности: `coin platform validate`.

## Build context

Корень `agents/` (Dockerfile paths в catalog относительно этой папки).
