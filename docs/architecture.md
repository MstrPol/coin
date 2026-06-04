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
Jenkins job coin-cli  →  go build  →  zip  →  Nexus Maven
                                              coin/platform/coin-cli/<ver>/coin-cli-<ver>-linux-<arch>.zip
                                              │
                         coin-lib (service pipeline)  →  fetch по GP profile → PATH
```

Репозиторий: `maven-releases` (release) или `maven-snapshots` (версия `*-SNAPSHOT`).

**Продукт не pin'ит версию CLI.** Единственная зависимость — `coin.template` + `coin.templateVersion`.  
Версия CLI и rev agent image задаются в **`golden-paths/<tpl>/<ver>/profile.yaml`** (platform bundle):

```yaml
agent:
  stack: go
  runtime: { go: "1.22" }
  rev: 2
coinCli:
  version: "0.0.0-SNAPSHOT"
```

- patch/minor CLI (совместимо) → bump `coinCli.version` в том же `vN` profile;
- major / breaking → новый каталог GP (`v2/`) + миграция `templateVersion` в проекте.

`platform.yaml` → `coinCli.min` — глобальный пол; `nexus.mavenBase` — для bootstrap в coin-lib.

Agent image содержит только toolchain (+ curl/unzip для bootstrap CLI).

## Выбор agent image (coin-lib)

```
coin.template + templateVersion  →  profile.yaml  →  agent.stack, runtime, rev
jenkins.runtime / jenkins.agent.image  →  optional override (проект)
agents/catalog.yaml  →  image ref по stack/runtime
```

Реализация: `ProfileLoader.groovy`, `StackImages.groovy`, `CoinCli.groovy`, `PodTemplate.groovy`.

## Как Jenkins вызывает CLI

`coinPipeline.groovy`:

```groovy
// 1. Лёгкий agent: checkout + resolve stack image (GP profile bundle)
// 2. K8s pod (jnlp + stack container)
// 3. В stack container:
stage('Bootstrap') { fetch coin CLI по profile.coinCli.version }
stage('Validate')  { sh 'coin validate --min-version <profile>' }
stage('Test')      { sh 'coin run test' }
stage('Build')     { sh 'coin run build' }
stage('Publish')   { withCredentials { sh 'coin run publish' } }
```

Groovy: pod template, credentials, стадии. Вся логика стадий — в GP scripts + coin-cli.

## Platform CI (monorepo `coin`)

| Артефакт | Jenkins job | Jenkinsfile |
|----------|-------------|-------------|
| coin CLI | `coin-cli` | `coin-cli/Jenkinsfile` |
| agent images | `coin-agents` | `coin-jenkins-agents/Jenkinsfile` |

Сервисные репозитории — отдельный multibranch job с `Jenkinsfile` + `coinPipeline()`.

Локальный стенд: [docker/README.md](../docker/README.md).

## Хранилище бинарей CLI (Maven)

```
<nexus>/repository/maven-releases/coin/platform/coin-cli/<version>/coin-cli-<version>-linux-amd64.zip
<nexus>/repository/maven-releases/coin/platform/coin-cli/<version>/coin-cli-<version>-linux-arm64.zip
```

Внутри zip — один файл `coin`. Classifier: `linux-amd64`, `linux-arm64`.

## Совместимость lib ↔ CLI

Версия CLI для продукта — **`profile.yaml` → `coinCli.version`**.  
`platform.yaml` → `coinCli.min` — глобальный пол (coin-lib проверяет до fetch).  
`coin validate --min-version` сверяет бинарь в PATH с pin из profile.

## Связанные документы

- [agent-build-model.md](agent-build-model.md)
- [golden-paths.md](golden-paths.md)
- [jenkins-setup.md](jenkins-setup.md)
- [config.md](config.md)
