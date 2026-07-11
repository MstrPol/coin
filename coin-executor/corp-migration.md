# Руководство по миграции coin-executor в корпоративную инфраструктуру

Данный документ описывает узкие места и настройки, которые необходимо адаптировать при переносе сборки `coin-executor` в защищенную корпоративную (Corp) среду.

## 1. Сборка бинарника (Jenkinsfile)

### 1.1 Инструменты (Jenkins Tools)
Для сборки бинарника `coin-executor` используется инструмент Go.
- В пайплайне: `tools { go 'golang-1.22' }`.
- В корпоративном Jenkins имя этого инструмента может отличаться (например, `Go-1.22`). Если оно отличается, скорректируйте его в `Jenkinsfile` или настройте алиас в самом Jenkins.

### 1.2 Приватные пакеты и Учетные данные (GOPROXY)
По умолчанию параметры указывают на публичный прокси:
- `GOPROXY` = `'https://proxy.golang.org,direct'`
- `GO_AUTH_CRED_ID` = `'go-auth-token'`

**При миграции:**
Переопределите параметр `GOPROXY` в настройках джобы, чтобы он указывал на корпоративный Nexus (например, `https://nexus.corp.local/repository/go-proxy/`). 
Если ваш Nexus требует авторизации, убедитесь, что секрет с ID `go-auth-token` существует в Jenkins. 
Пайплайн настроен так, что он:
1. Автоматически извлекает секрет.
2. Подменяет `~/.netrc`.
3. Отключает валидацию SSL (переменными `GOINSECURE="*"` и `GIT_SSL_NO_VERIFY=true`), чтобы избежать падений на корпоративных (MITM) сертификатах.

## 2. Публикация в Nexus
Функция публикации бинарника (`publishExecutor`) теперь встроена прямо в `Jenkinsfile`.
Вам потребуются следующие переменные/секреты в корпоративной среде (уже пробрасываются в Jenkinsfile, просто убедитесь, что они существуют):
- `NEXUS_USER` и `NEXUS_PASSWORD` (через Credentials ID `nexus-admin`).
- `COIN_API_KEY` (через Credentials ID `coin-publisher-api-key`).
- `COIN_API_URL` (должен указывать на корпоративный эндпоинт `coin-api`, например `http://coin-api.coin.svc.cluster.local:8090`).

## 3. Образ Агента (Dockerfile.agent)
Если вы собираете Docker-образ агента (`Dockerfile.agent`) с упакованным `coin-executor`:
- Бинарник теперь собирается прямо внутри `Dockerfile.agent` (stage `executor-builder`), поэтому при сборке образа агента тоже необходимо пробрасывать переменные `GOPROXY`, отключать SSL-валидацию (`GOINSECURE="*"`, `GIT_SSL_NO_VERIFY="true"`) и монтировать секрет авторизации (`--secret id=go_auth`).
- `Dockerfile.agent` поддерживает `ARG REGISTRY=""` для подмены хаба.
- При сборке этого образа внутри корпоративной сети передавайте аргумент `--opt build-arg:REGISTRY=artifactory.corp.local/`, чтобы образы `golang`, `moby/buildkit` и `jenkins/inbound-agent` брались из вашего проксирующего реестра.
