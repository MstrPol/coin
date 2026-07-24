# ADR: BuildKit secrets (`nexus` / `osc`)

## Статус

accepted (2026-07-24)

**Связанные:** [build-engine-contract](build-engine-contract.md), [coin-ci-runtime](coin-ci-runtime.md), [jenkins-lib-http-nexus](jenkins-lib-http-nexus.md), [config.md — credentials](../config.md#контракт-jenkins-credentials-platform).

## Контекст

Corp-сборка внутри managed Containerfile нуждается в:

- auth к Maven/raw Nexus (`jenkins.credentials.nexus`);
- corp HTTP(S) proxy OSC (`jenkins.credentials.osc`).

Секреты нельзя запекать в слои образа и нельзя полагаться только на env агента: сеть `RUN` идёт в network namespace BuildKit worker, не в shell Jenkins.

Выбрано: **BuildKit `RUN --mount=type=secret`** + Jenkins `withCredentials` → файлы на агенте → `coin-executor` передаёт `--secret` в `buildctl`.

## Решение

### Поток

```mermaid
sequenceDiagram
  participant J as coin-lib (Jenkins)
  participant FS as .coin/secrets/
  participant E as coin-executor
  participant BK as buildctl / buildkitd
  participant CF as Containerfile RUN

  J->>J: withCredentials(nexus, osc)
  J->>FS: materialize files (0600)
  J->>E: run task (build / containerfile)
  E->>BK: buildctl … --secret id=…,src=…
  BK->>CF: mount /run/secrets/<id>
  Note over CF: читать secret; не COPY в образ
  J->>FS: cleanup (finally)
```

### Стабильные BuildKit secret IDs (контракт GP)

| Secret id | Содержимое файла | Источник Jenkins |
|-----------|------------------|------------------|
| `coin-nexus-user` | username (plain text, одна строка) | `jenkins.credentials.nexus` (usernamePassword) |
| `coin-nexus-password` | password | `jenkins.credentials.nexus` |
| `coin-osc-user` | username | `jenkins.credentials.osc` (usernamePassword) |
| `coin-osc-password` | password | `jenkins.credentials.osc` |

### Platform build-args (не секреты)

Стабильные `ARG` для GP Containerfile. Executor передаёт через  
`buildctl --opt build-arg:NAME=value`. В ARG/слое — **только URL без userinfo**.

| Build-arg | Назначение | SoT значения |
|-----------|------------|--------------|
| `COIN_CORP_PROXY_URL` | Base URL corp HTTP(S) proxy (OSC), без credentials | `coin.corpProxyUrl` из merged lib defaults (пустая строка на local) |
| `COIN_NEXUS_URL` | Base URL artifact/deps registry (Maven/raw/Go proxy и т.п.), куда смотрят зависимости и куда публикуется внутренняя разработка | `manifest.destinations[]`: **первый** элемент с непустым `artifactRepositoryBase` |

Если подходящего destination нет — `COIN_NEXUS_URL` не передаётся (пустой / omit).  
Несколько destinations с `artifactRepositoryBase`: пока правило «первый в массиве»; явный pin по `id` — follow-up при появлении multi-artifact use case.

Containerfile собирает auth URL сам: secrets (`coin-nexus-*` / `coin-osc-*`) + эти ARG.

Default mount path в Dockerfile frontend: `/run/secrets/<id>` (можно `target=`).

### Layout на агенте

```
.coin/secrets/           # gitignore; chmod 0700
  coin-nexus-user
  coin-nexus-password
  coin-osc-user
  coin-osc-password
```

- Пишет **только coin-lib** (Jenkins Credentials binding).
- Читает **только coin-executor** для `buildctl --secret`.
- Product / GP / Git — никогда.

### Роли

| Компонент | Ответственность |
|-----------|-----------------|
| **coin-lib** | `withCredentials` по merged `jenkins.credentials.nexus` / `osc`; запись файлов; `chmod`; `finally` delete; обернуть stages, где возможен BuildKit network (`containerfile` run, `build`, publish rebuild) |
| **coin-executor** | Secrets из `.coin/secrets/` → `--secret`; build-arg `COIN_CORP_PROXY_URL` / `COIN_NEXUS_URL` из cfg + manifest; **не** логировать секреты |
| **GP Containerfile** | `ARG COIN_*` + явный `RUN --mount=type=secret,id=coin-…`; dual local/corp — `required=false` где уместно |
| **Product config** | Только override credential **IDs**; не пути, не mount |

### Пример (GP / corp)

```dockerfile
# syntax=docker/dockerfile:1.8
ARG COIN_CORP_PROXY_URL
ARG COIN_NEXUS_URL
RUN --mount=type=secret,id=coin-osc-user \
    --mount=type=secret,id=coin-osc-password \
    --mount=type=secret,id=coin-nexus-user \
    --mount=type=secret,id=coin-nexus-password \
    set -eu; \
    OU=$(cat /run/secrets/coin-osc-user); \
    OP=$(cat /run/secrets/coin-osc-password); \
    export HTTPS_PROXY="http://${OU}:${OP}@${COIN_CORP_PROXY_URL#http://}"; \
    export HTTP_PROXY="$HTTPS_PROXY"; \
    NU=$(cat /run/secrets/coin-nexus-user); \
    NP=$(cat /run/secrets/coin-nexus-password); \
    # пример: Go module proxy / Maven — URL из ARG, auth из secrets
    export GOPROXY="${COIN_NEXUS_URL}"; \
    # … настроить netrc/settings.xml из NU/NP при необходимости …
    go mod download
```

Local sample Containerfiles **без** mount / без зависимости от ARG — как сейчас; secrets и build-arg на агенте могут быть unused.

### `buildctl` (executor)

**Secrets** — для каждого существующего файла в `.coin/secrets/` с именем ∈ контрактной таблицы:

```text
--secret id=coin-osc-user,src=.coin/secrets/coin-osc-user
```

Отсутствующий файл → секрет **не** передаётся (local pilot без OSC).  
Containerfile с `required=true` (default) без секрета → fail на `RUN` (ожидаемо для corp GP на стенде без creds).

**Build-args** (если значение непустое):

```text
--opt build-arg:COIN_CORP_PROXY_URL=http://proxy.corp.example:8080
--opt build-arg:COIN_NEXUS_URL=http://nexus:8081/repository/maven-releases
```

### Podman fallback

Primary path — BuildKit. Если сработал podman fallback и есть `.coin/secrets/*`:

- **предпочтительно:** `podman build --secret id=…,src=…` (Podman 4+);
- иначе **hard fail** с сообщением: build secrets требуют BuildKit.

Не молча собирать без секретов.

### Безопасность

- `.coin/secrets/` в `.gitignore` (platform + product templates).
- Не `COPY` secrets в любой stage.
- Не печатать значения в Jenkins / executor logs (допустимо: список id).
- Cleanup в `finally` даже при fail stage.
- Cache: secret mount не должен попадать в cache key содержимого (BuildKit исключает secret из cache checksum по design); GP не пишет secret в слой.

### Local pilot

- Credentials `osc-proxy` / использование `nexus` для build — **опциональны**.
- Defaults ключей остаются в `coin-lib-defaults.yaml`.
- Materialize: если credential ID нет в Jenkins → skip файла + warn (не валить пайп на local).
- Corp: отсутствие обязательного secret при `required` mount → fail build.

## Последствия

Плюсы:

- стандартный Docker/BuildKit контракт;
- секреты не в слоях и не в Git;
- чёткое разделение glue (lib) / build (executor) / template (GP).

Минусы / стоимость:

- corp GP Containerfile обязан объявлять mount;
- lib оборачивает build-related stages в credentials + IO;
- podman path усложняется или запрещается при secrets.

## Отклонённые альтернативы

| | Почему нет |
|---|------------|
| Только env на агенте (`HTTP_PROXY`) | Не гарантирует видимость в BuildKit worker network |
| Prefetch deps на агенте + `COPY` | Ломает managed Containerfile / cache model |
| Секреты в manifest / GP release | Нарушение SoT; credentials — Jenkins-only |
| Один blob `user:pass@url` как единственный id | Хуже для Maven `settings.xml` / раздельных toolchains; URL прокси лучше non-secret |

## Implementation checklist (follow-up, не этот ADR)

1. `.gitignore`: `.coin/secrets/`
2. `coin-lib`: materialize secrets + wrap build/test/publish; `coin.corpProxyUrl` in defaults
3. `coin-executor` `build.Options` / `buildkitArgs`: `--secret` из `.coin/secrets/`; `--opt build-arg:COIN_CORP_PROXY_URL` / `COIN_NEXUS_URL` ✅ (2026-07)
4. Resolve `COIN_NEXUS_URL`: первый `destinations[]` с непустым `artifactRepositoryBase` ✅
5. Tests: secrets/build-args when present; omit when absent ✅
6. Corp GP template example + docs how-to ✅ → [how-to/containerfile-build-secrets.md](../how-to/containerfile-build-secrets.md)
7. JCasC seed `osc-proxy` на corp/local when ready ⏳

## Уточнения (приняты)

| # | Тема | Решение |
|---|------|---------|
| U1 | Тип Jenkins cred для OSC | usernamePassword (`osc-proxy`); host в `coin.corpProxyUrl` |
| U2 | Build-args URL | `COIN_CORP_PROXY_URL` ← `coin.corpProxyUrl`; `COIN_NEXUS_URL` ← первый `artifactRepositoryBase` |
| U3 | Materialize span | На любой `buildctl`; предпочтительно один wrap на `container('builder')` |
| U4 | Docker registry в `RUN` | Не через этот ADR — agent `~/.docker/config.json` / buildkit session |
