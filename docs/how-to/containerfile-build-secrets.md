# How-to: secrets и build-arg в managed Containerfile

Контракт: [ADR buildkit-secrets](../adr/buildkit-secrets.md).

Платформа **всегда** (если значения есть) передаёт в `buildctl`:

- secrets: `coin-nexus-user`, `coin-nexus-password`, `coin-osc-user`, `coin-osc-password`
- build-arg: `COIN_CORP_PROXY_URL`, `COIN_NEXUS_URL`

Автор Containerfile **сам** решает, объявлять ли `ARG` / `--mount`.

## Каталог id / ARG

| Имя | Тип | Содержимое |
|-----|-----|------------|
| `coin-nexus-user` | secret mount | username Nexus |
| `coin-nexus-password` | secret mount | password Nexus |
| `coin-osc-user` | secret mount | username OSC-прокси |
| `coin-osc-password` | secret mount | password OSC-прокси |
| `COIN_NEXUS_URL` | build-arg | base URL deps registry (без userinfo) |
| `COIN_CORP_PROXY_URL` | build-arg | base URL proxy (без userinfo) |

Путь mount по умолчанию: `/run/secrets/<id>`.

## Пример

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
    # deps URL из ARG; auth — из secrets (netrc / settings.xml / GOPROXY+header — по стеку)
    echo "nexus=${COIN_NEXUS_URL}"; \
    go mod download
```

Local samples могут **не** использовать mount/ARG — тогда runtime всё равно может прокинуть args/secrets unused.

## Что не делать

- Не `COPY` secrets в образ
- Не класть пароли в `ARG` / `ENV`
- Не придумывать свои secret id — платформа их не доставит
