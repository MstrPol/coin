# Руководство по миграции coin-api в корпоративную инфраструктуру

Данный документ описывает узкие места и настройки, которые необходимо будет адаптировать при переносе CI/CD процессов и деплоя модуля `coin-api` в защищенную корпоративную (Corp) среду.

## 1. Сборка и Jenkinsfile

### 1.1 Registries и Учетные данные
В `coin-api/Jenkinsfile` используются дефолтные параметры для локальной среды:
- `SOURCE_REGISTRY` (default: `'registry.local'`)
- `TARGET_REGISTRY` (default: `'registry.local:5000'`)
- `SOURCE_CRED_ID` (default: `'source-registry-creds'`)
- `TARGET_CRED_ID` (default: `'target-registry-creds'`)
- `GOPROXY` (default: `'https://proxy.golang.org,direct'`)
- `GO_AUTH_CRED_ID` (default: `'go-auth-token'`)

**При миграции:**
Вам потребуется изменить значения по умолчанию в самом `Jenkinsfile` или задавать их явно на уровне настройки Job'ы в корпоративном Jenkins.
*Пример:* `TARGET_REGISTRY = 'artifactory.corp.local/coin-images'` и `GOPROXY = 'https://nexus.corp.local/repository/go-proxy/'`.
Убедитесь, что ID секретов (`CRED_ID`, включая `GO_AUTH_CRED_ID` для токена Nexus) существуют в хранилище корпоративного Jenkins.

### 1.2 Названия инструментов (Jenkins Tools)
В пайплайне жестко зафиксировано использование инструмента `go`:
```groovy
tools {
    go 'golang-1.25'
}
```
**При миграции:**
В корпоративном Jenkins имя инструмента может отличаться (например, `Go-1.25` или `golang-corp`). Проверьте раздел *Global Tool Configuration* в Jenkins и укажите правильное имя, чтобы сборка не упала на стейдже тестирования.

---

## 2. Docker-образ (Dockerfile)

### 2.1 Подмена Base Images (ARG REGISTRY)
Для того чтобы `Dockerfile` не тянул базовые образы напрямую из DockerHub (что запрещено в большинстве corp-сетей), мы внедрили `ARG REGISTRY=""` на уровне сборки.

**При миграции:**
Вам не нужно изменять `Dockerfile`. Достаточно убедиться, что `buildctl` в `Jenkinsfile` передает корректный `--opt build-arg:REGISTRY=${params.SOURCE_REGISTRY}/`.
Это автоматически преобразует конструкцию:
`FROM ${REGISTRY}golang:1.25-alpine` -> `FROM artifactory.corp.local/docker-proxy/golang:1.25-alpine`

### 2.2 Корпоративные Root CA (Опционально)
Если `coin-api` будет делать HTTPS-запросы к внутренним корпоративным ресурсам (например, Gitea или Nexus) с самоподписанными сертификатами, потребуется добавить их в Docker-образ:
*Возможное изменение Dockerfile:*
```dockerfile
COPY corp-ca.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
```

---

## 3. Helm-чарт и Деплой (values.yaml)

При переносе Helm-чарта в корпоративный кластер K8s, обратите внимание на следующие параметры:

### 3.1 Инфраструктурные зависимости
Адреса базы данных и Nexus, скорее всего, изменятся:
- `config.dbHost`: Изменить с внутреннего `coin-pg-cluster.db.svc` на корпоративный инстанс (например, `pg-coin.db.corp.local`).
- `config.nexusUrl`: Изменить на корпоративный адрес Artifactory/Nexus.

### 3.2 Ingress и Сертификаты
В корпоративной среде доступ "снаружи" к `coin-api` будет жестче регулироваться:
- Потребуется добавить **специфичные корпоративные аннотации Ingress** (например, для Nginx/HAProxy Ingress Controller, интеграции с Keycloak/OIDC).
- Необходимо настроить блок `tls`, указав существующий `secretName`, который генерируется вашим корпоративным Cert-Manager'ом.

### 3.3 Секреты (ImagePullSecrets)
Убедитесь, что в пространстве имен (namespace), куда деплоится модуль, создан секрет типа `kubernetes.io/dockerconfigjson` с учетными данными для пулла образов из корп. Registry, и он корректно указан в `values.yaml` в поле `image.pullSecret` или `imagePullSecrets`.

---

## 4. Использование Podman вместо Docker
В корпоративном CI строго запрещено использование классического Docker Daemon.
- Для сборки образов уже настроен и используется `buildctl`.
- Если в будущем для пайплайна или E2E-тестов модуля `coin-api` понадобится запуск временных контейнеров (БД, mock-серверы), используйте **исключительно** команды `podman` или `podman-compose`. Вызовы `docker run` или монтирование `docker.sock` в среде упадут с ошибкой.
