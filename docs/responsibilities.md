# Разделение ответственности (каркас)

Этот документ фиксирует **границы владения** между командами разработки и DevOps/Platform (Coin).
Цель — уменьшить размывание ответственности, ускорить поддержку L1/L2 и сделать возможными обязательные корпоративные изменения (QG, версионирование, security) без “100 PR во все репозитории”.

## Принцип

- **Проект** описывает *что именно он делает* (команды теста/сборки/паблиша, параметры сборки).
- **Coin** описывает *как именно это запускается в компании* (оркестрация Jenkins, агенты/образы, секреты, QG, политика версий).

## Что управляет разработчик в своём проекте

- **Код и зависимости**: `pyproject.toml` / `pom.xml` / `package.json` / `go.mod`.
- **Расширения CI-команд** в `.coin/config.yaml`.
  - По умолчанию test/build/publish выполняются стандартными сценариями из Coin.
  - Команда может добавить `preCommands`/`postCommands` или полностью заменить stage через `commands`.
  - Команда отвечает за корректность своих дополнительных команд.
- **Параметры контейнерной сборки** (если `pipeline.build.target: container`):
  - команда передаёт параметры (порт, command/entrypoint, аргументы),
  - команда **не управляет Dockerfile напрямую**: Dockerfile централизован в Coin.
- **Локальные флаги пайплайна** в `.coin/config.yaml`:
  - включить/выключить `test/build/publish` (в пределах разрешённого),
  - выбрать `build.target: package|container`,
  - указать дополнительные команды `pipeline.*.preCommands/postCommands` (если разрешено политикой).

## Что управляет DevOps/Platform через Coin

- **Оркестрация в Jenkins**: multibranch, K8s pod template, логирование, стандартные stage’и.
- **CI‑образы (toolchain)**: `coin-images/*` (python/jdk/node/go и т.п.), политика обновлений.
- **Registry/credentials/secrets**: naming, креды, Vault/Jenkins credentials, запрет секретов в коде.
- **Корпоративные инварианты** (обязательные):
  - **Модель версионирования ПО** (единая для всех стеков) и способ инжекта версии в сборку.
    - версия вычисляется в Coin (`COIN_VERSION`) из Git/Jenkins;
    - Gradle/Maven/uv/Go tooling не являются источником версии;
    - плагины и настройки версионирования в проектах не используются как корпоративный механизм.
  - **Quality Gates**: что обязано выполняться (Sonar, SAST/DAST, лицензии и т.д.), пороги и включение по датам.
  - **Security policies**: базовые образы allowlist, SBOM, подпись, сканирование, запреты.
  - **Политика релиза**: какие теги релизные, approvals, promotion, “build once deploy many”.
- **Централизованные Dockerfile** (если включено в компании):
  - Dockerfile‑шаблоны живут в `coin` и версионируются,
  - Coin генерирует/подкладывает Dockerfile и `.dockerignore` в workspace при сборке,
  - проект не хранит Dockerfile; наличие `Dockerfile` в репозитории сервиса считается нарушением контракта.
- **Политика обязательных обновлений**:
  - минимальная допустимая версия golden path (`coin.templateVersion` ≥ `COIN_MIN_TEMPLATE_VERSION`),
  - запрет запуска без pinning (template+version) — при включённой политике.

## Что запрещено/ограничено в проекте (чтобы не было “обходов”)

Ровно набор ограничений зависит от режима компании, но каркас такой:

- **Нельзя менять модель версионирования** (если `coin.versioning.mode: corporate`).
- **Нельзя отключать обязательные QG** (если они policy‑enforced в Coin).
- **Нельзя хранить shell-скрипты CI как часть golden path** без необходимости: стандартные сценарии ведёт Coin.
- **Нельзя хранить секреты** в `.coin/config.yaml`/переменных репо — только через Jenkins/Vault.
- **Нельзя хранить Dockerfile** в репозитории сервиса при `pipeline.build.target: container`.
- **Не нужно хранить `.dockerignore`** ради CI: managed `.dockerignore` генерирует Coin.

## Артефакты и “кто владеет чем”

| Артефакт | Где | Владелец |
|---------|-----|----------|
| `coin-lib` | monorepo `coin` | DevOps/Platform |
| `coin-images/*` | monorepo `coin` | DevOps/Platform |
| `coin-templates/*` | monorepo `coin` | DevOps/Platform (эталон) |
| `coin-lib/resources/dockerfiles/*` | monorepo `coin` | DevOps/Platform |
| `coin-lib/resources/dockerignore/*` | monorepo `coin` | DevOps/Platform |
| `coin-lib/resources/scripts/*` | monorepo `coin` | DevOps/Platform |
| `.coin/config.yaml` | репо сервиса | команда (но с enforced policy) |
| `Dockerfile` runtime | генерируется Coin в CI | DevOps/Platform |

## Как поддержка отличает “сломали в проекте” от “сломали платформу”

Минимальный стандарт диагностики:

1. В логе Coin печатает: `coin-lib@X`, `template@Y`, `stackImage`, `COIN_VERSION`.
2. `Coin Validate` (перед стадиями) проверяет:
   - schema config,
   - допустимость project overrides (`commands`, `preCommands`, `postCommands`),
   - pinning шаблона,
   - минимальную версию шаблона.

Если Validate не проходит — это **проблема проекта/контракта**.
Если Validate проходит, но падает pod/образ/credentials/QG — это **platform**.

