# Архитектура Coin

## Принцип разделения

```
┌─────────────────────────────────────────────────────────┐
│  Jenkins (coin-lib)  — ТОЛЬКО оркестрация               │
│  выбор агента, credentials, стадии, QG-пороги           │
├─────────────────────────────────────────────────────────┤
│  coin CLI (в образе агента)  — ВСЯ логика               │
│  валидация, версионирование, сборка, публикация          │
└─────────────────────────────────────────────────────────┘
```

Jenkins **не вычисляет** версии, **не рендерит** Dockerfile, **не знает** о деталях стеков.
Он вызывает `coin` и получает результат.

## Компоненты

| Компонент | Язык | Назначение |
|-----------|------|------------|
| `coin-lib` | Groovy | Тонкий Jenkins-оркестратор |
| `coin-cli` | Go | CLI: validate, version, run, dockerfile render |
| `coin-jenkins-agents` | Docker | CI agent images (toolchain + coin CLI + pack) |
| `coin-golden-paths` | — | Golden paths: profile, scripts, runtime-only Dockerfile |
| `coin-starters` | — | Скелетоны новых репозиториев |

## Модель сборки

**Native compile в agent → runtime-only Dockerfile → registry.**

Подробно — [agent-build-model.md](agent-build-model.md).

```
coin run test   → native в stack container
coin run build  → native compile → docker/kaniko pack
coin run publish → push app image
```

Managed Dockerfile GP — **без** builder stage. Agent image — единственная CI-среда для test и compile.

## coin CLI

Go-бинарь. Golden paths загружаются из `coin-golden-paths/` (local или Nexus tarball). См. [golden-path-versioning.md](golden-path-versioning.md).

### Команды

```
coin validate                                  # валидация .coin/config.yaml
coin init                                      # новый проект (визард)
coin version                                   # текущая версия (read-only)
coin version bump patch|minor|major            # snapshot-тег
coin version bump patch|minor|major --type rc  # RC-тег (release/*)
coin run test|build|publish                    # стадии CI
coin dockerfile render                         # .coin/generated/Dockerfile (runtime-only)
```

### Локальный запуск

```bash
coin validate
coin version
coin run test
```

Требуется toolchain стека на host (go, uv, …) и `COIN_GOLDEN_PATHS_DIR` при работе вне agent image.

## Доставка coin CLI в agent

```
Jenkins job coin-cli  →  go build  →  Nexus raw/coin-cli/<ver>/coin_linux_<arch>
                                              │
                         coin-lib (service pipeline)  →  curl/wget → PATH
```

Agent image содержит только toolchain + GP; версия CLI — `images.yaml` → `coinCli.min`.

## Выбор agent image (coin-lib)

```
coin.template  →  catalog.yaml  →  stack
profile.agent.runtime  →  дефолт версии (GP)
jenkins.runtime  →  optional override (проект)
images.yaml stacks[stack][version]  →  image ref
```

Реализация: `StackImages.groovy` + `PodTemplate.groovy`.

## Как Jenkins вызывает CLI

`coinPipeline.groovy`:

```groovy
// 1. Лёгкий agent: checkout + resolve stack image
// 2. K8s pod (jnlp + stack container)
// 3. В stack container:
stage('Validate') { sh 'coin validate' }
stage('Test')     { sh 'coin run test' }
stage('Build')    { sh 'coin run build' }
stage('Publish')  {
    withCredentials([...]) { sh 'coin run publish' }
}
```

Groovy: pod template, credentials, стадии. Вся логика стадий — в GP scripts + coin-cli.

## Platform CI (monorepo `coin`)

| Артефакт | Jenkins job | Jenkinsfile |
|----------|-------------|-------------|
| coin CLI | `coin-cli` | `coin-cli/Jenkinsfile` |
| agent images | `coin-agents` | `coin-jenkins-agents/Jenkinsfile` |

Сервисные репозитории — отдельный multibranch job с `Jenkinsfile` + `coinPipeline()`.

Локальный стенд: [docker/README.md](../docker/README.md).

## Хранилище бинарей CLI

```
<nexus>/repository/coin-cli/<version>/coin_linux_amd64
<nexus>/repository/coin-cli/<version>/coin_linux_arm64
```

## Совместимость lib ↔ CLI

`images.yaml` → `coinCli.min`. `coin validate --min-version` проверяет версию бинаря в PATH.

## Связанные документы

- [agent-build-model.md](agent-build-model.md)
- [golden-paths.md](golden-paths.md)
- [jenkins-setup.md](jenkins-setup.md)
- [config.md](config.md)
