# Архитектура Coin (Control Plane v2)

## Принцип разделения

```
┌─────────────────────────────────────────────────────────┐
│  Jenkins (тонкий Jenkinsfile)  — resolve + pod + creds  │
├─────────────────────────────────────────────────────────┤
│  coin-executor  — validate, run stages, report          │
├─────────────────────────────────────────────────────────┤
│  coin-api + PostgreSQL + Nexus  — manifest, GP content, registry    │
├─────────────────────────────────────────────────────────┤
│  Nexus content/  — published scripts, Dockerfile, orchestration     │
└─────────────────────────────────────────────────────────┘
```

Jenkins **не** checkout'ит platform@main, **не** pin'ит CLI, **не** читает profile.yaml.

## Компоненты

| Компонент | Назначение |
|-----------|------------|
| `coin-api` | Resolve manifest, registry metadata, GP authoring (PG) |
| `coin-executor` | Runtime pipeline (fetch content by URL) |
| `coin-jenkins-agents/` | CI agent images (`ci-go`, …) |
| `coin-starters/` | Скелетоны репозиториев + thin `Jenkinsfile.coin` |

## Модель сборки

Native compile в agent → runtime-only Dockerfile → registry.  
Подробно — [agent-build-model.md](agent-build-model.md).

```
coin-executor run --stage test    → GP script test.sh
coin-executor run --stage build   → native compile + pack-image.sh
coin-executor run --stage publish → GP script (when: tag)
```

## Продуктовый контракт

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

Strict v2 — поля `template` / `templateVersion` **не** поддерживаются.  
См. [config.md](config.md).

## Доставка executor

```
Jenkins job coin-executor  →  go build  →  Nexus raw coin-executor/{ver}/coin-executor-linux-{arch}
                                              │
Product pipeline (Bootstrap)  →  curl manifest.executor.url
```

## Dynamic agent

Образ из `manifest.runtime.image`. Pod template — inline в `Jenkinsfile.coin`.

Локальный стенд: после `docker compose up` выполнить `make endpoints` — k3s Endpoints для jenkins/nexus/gitea.

## Platform CI

| Артефакт | Jenkins job |
|----------|---------------|
| coin-executor | `coin-executor` |
| agent images | `agents-build` |

## Связанные документы

- [control-plane.md](control-plane.md)
- [config.md](config.md)
- [jenkins-setup.md](jenkins-setup.md)
- [docker/README.md](../docker/README.md)
