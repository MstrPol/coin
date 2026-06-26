---
name: cicd
description: >-
  Создаёт и ревьюит Jenkinsfile и документацию helm-chart-requirements.md для модулей проекта.
  Обеспечивает сборку через BuildKit (buildctl), авторизацию в нескольких registries, фокус на Groovy вместо bash и адаптивность под corp-среду.
---

# Разработка CI/CD для модулей (Jenkinsfile & Helm)

Этот скилл применяется при создании или обновлении пайплайнов (`Jenkinsfile`) и требований к деплою (`helm-chart-requirements.md`) для отдельных модулей инфраструктуры (например, `coin-api`, `coin-ui` и т.д.).

## 🏗 Ключевые требования к Jenkinsfile

1. **Максимум Groovy, минимум Bash**
   - Вся логика (вычисление тегов, проверка условий, парсинг манифестов, настройка параметров) должна быть реализована средствами Jenkins Pipeline (Groovy).
   - Шаг `sh` используется **только** для прямого вызова внешних утилит (например, `buildctl`), но не для скриптования логики.
   - Существующие в репозитории `.sh` скрипты **не используются** в Jenkinsfile (они предназначены исключительно для локального запуска).

2. **Сборка через BuildKit (`buildctl`) и полный отказ от Docker**
   - Использование классического Docker (Docker Daemon, `docker build`, `docker push`, сокет `docker.sock`) **категорически запрещено**.
   - Для сборки и публикации образов используется **только** `buildctl` (BuildKit CLI).
   - В `buildctl` нужно корректно прокидывать контекст сборки, путь к `Dockerfile` и аргументы сборки.

3. **Запуск контейнеров только через Podman**
   - Так как `buildctl` предназначен только для сборки, а Docker запрещен, любая необходимость запуска контейнеров в CI (например, для E2E-тестов, поднятия временной БД) должна реализовываться исключительно через `podman` или `podman-compose`.

3. **Авторизация в нескольких Registries**
   - Пайплайн должен уметь забирать базовые образы из одного registry (например, зеркало / proxy), а публиковать готовый образ в другой registry (release / snapshot).
   - Реализуется через формирование `config.json` для Docker или через плагин Docker Credentials, с поддержкой нескольких URL и Credentials ID.

4. **Портируемость (Local → Corp environment)**
   - Пайплайн не должен содержать захардкоженных URL-адресов.
   - Все registry URL, Credentials ID, K8s namespaces должны быть вынесены в `parameters` или `environment` переменные. Это позволит без изменения кода перенести пайплайн в корпоративную инфраструктуру.

5. **Локализация в директории модуля**
   - Вся логика работы пайплайна должна выполняться внутри директории целевого модуля.
   - Папка `docker/` в корне проекта **не используется** и игнорируется при сборке модулей.

6. **Отказ от Kubernetes-агентов**
   - В `Jenkinsfile` **строго запрещено** использовать Kubernetes-агентов (`agent { kubernetes { ... } }`).
   - Используй стандартных агентов (например, `agent any` или специфичный лейбл `agent { label 'docker' }`).

7. **Сборка и тестирование модуля**
   - При необходимости пайплайн должен включать отдельные стадии (stages) для компиляции и тестирования кода модуля перед сборкой Docker-образа (например, вызов `go test`, `npm test` и т.д.).

8. **Использование Jenkins Tools**
   - Инструменты сборки (Go, Node.js, JDK и др.) должны подключаться через блок `tools { ... }`, чтобы не зависеть от глобально установленных пакетов на агенте.

9. **Авторизация пакетных менеджеров (npm, go, pip)**
   - Пайплайн обязан безопасно прокидывать секреты/токены для инструментов сборки, чтобы они могли скачивать зависимости из корпоративных приватных реестров (Nexus).
   - В `Jenkinsfile` секрет извлекается через `withCredentials([string(credentialsId: '...', variable: 'TOKEN')])` и передается в `buildctl` флагом `--secret id=token_name,env=TOKEN`.
    - В `Dockerfile` токен потребляется через `RUN --mount=type=secret,id=token_name`, а URL реестра передается через `--opt build-arg`. Использование `ARG` для токенов запрещено.

10. **Авторизация при локальном тестировании (на агенте)**
    - Если в пайплайне есть шаги (stages), выполняющие тестирование или линтинг локально на Jenkins-агенте (до сборки Docker-образа), они также должны использовать `withCredentials` для подключения к приватным реестрам.
    - Перед вызовом команд типа `go test` или `npm ci` необходимо сконфигурировать авторизацию (например, сгенерировать `~/.netrc` для Go или сделать `npm config set` для Node.js), используя переданные секреты.

11. **Отключение проверки сертификатов (SSL/TLS)**
    - В корпоративных сетях часто используются прокси с подменой SSL-сертификатов (MITM), из-за чего стандартные инструменты падают с ошибкой проверки сертификата.
    - Для Go используйте переменные окружения `GOINSECURE="*"` и `GIT_SSL_NO_VERIFY=true`.
    - Для NPM используйте команду `npm config set strict-ssl false` и переменную окружения `NODE_TLS_REJECT_UNAUTHORIZED=0`.
    - Эти настройки должны применяться как при вызове команд локально на агенте (через `withEnv` в `Jenkinsfile`), так и пробрасываться через `ARG`/`ENV` в `Dockerfile` для `buildctl`.

12. **Структура Jenkinsfile**
    - Все кастомные функции (`def ...()`, `@NonCPS`) должны быть объявлены **в конце** файла `Jenkinsfile` (после блока `pipeline { ... }`), чтобы упростить чтение основной логики (stages) пайплайна.

## 🐳 Ключевые требования к Dockerfile

При написании или редактировании `Dockerfile` модулей:
1. **Обязательно соблюдай правила** скилла `.cursor/skills/dockerfile/SKILL.md`.
2. **Минималистичность**: только необходимые слои, оптимизация размера (через `--mount=type=cache`), использование distroless или `-slim` образов.
3. **Корпоративная переносимость**: используй `ARG` для registry базовых образов (например, `ARG REGISTRY=""` и `FROM ${REGISTRY}golang:1.22-bookworm`), чтобы в corp-сети можно было легко подменить источник без изменения кода файла.

## 📝 Требования к `helm-chart-requirements.md`

Для каждого модуля должен создаваться (или обновляться) файл `helm-chart-requirements.md` в корне модуля, описывающий **обязательные** требования к будущему Helm-чарту. Файл должен содержать:

1. **Управление образами и доступами**
   - Поддержка параметров `image.repository` и `image.tag`.
   - Поддержка `imagePullSecrets` для скачивания из приватных corp-registries.
2. **Ресурсы и масштабирование**
   - Блок `resources` (requests/limits для CPU и RAM) для обеспечения стабильности в кластере.
   - Требования к HPA (Horizontal Pod Autoscaler) (опционально, если модуль stateless).
3. **Конфигурация и окружение**
   - Конкретный список необходимых переменных окружения с описанием их назначения (зачем нужны, на что влияют).
   - Примеры YAML с реализацией проброса (куски `values.yaml` и `deployment.yaml` с `ConfigMap`, `Secret` или `env`).
   - Настройка портов и Service (внутренний трафик).
4. **Healthchecks**
   - Описание маршрутов или команд для `livenessProbe` и `readinessProbe` (с задержками и таймаутами).
5. **Ingress (если применимо)**
   - Настройка роутинга и TLS для веб-доступа (например, для `coin-api` или `coin-ui`).

## 📖 Требования к руководству по миграции (`corp-migration.md`)

Для обеспечения гладкого переноса модуля из локальной среды (Local/Gitea) в корпоративную (Corp/K8s/Nexus/Artifactory), для каждого модуля необходимо генерировать файл `corp-migration.md` в корне модуля.

В этом файле должно быть описано:
1. **Базовые образы и Registries**: Как переопределить переменные `SOURCE_REGISTRY` и `TARGET_REGISTRY` в `Jenkinsfile` или передать `ARG REGISTRY` в `Dockerfile`, чтобы образы тянулись из корпоративного кэша/прокси.
2. **Учетные данные (Credentials)**: Какие ID секретов (например, `SOURCE_CRED_ID`) используются в Jenkins и как их перенастроить под корпоративный Vault/Credentials Store.
3. **Jenkins Tools**: Имена инструментов в блоке `tools {}` (например, `go 'golang-1.25'`). В корпоративном Jenkins они могут называться иначе (например, `Go-1.25-Corp`).
4. **Helm и Инфраструктура**:
   - Адреса баз данных, брокеров сообщений и других зависимостей (нужно будет поменять в `values.yaml`).
   - Настройка Ingress (корпоративные аннотации, TLS-сертификаты).
   - Подключение корпоративных корневых сертификатов (Custom CA) в Dockerfile, если модулю необходимо делать HTTPS-запросы к внутренним ресурсам.

## Пример фрагмента Jenkinsfile (Сборка buildctl с multi-registry)

```groovy
pipeline {
    // Отказ от k8s-агентов. Используется стандартный агент или настроенный по лейблу.
    agent any

    tools {
        go 'go-1.25' // Пример использования инструмента Jenkins
    }

    parameters {
        string(name: 'SOURCE_REGISTRY', defaultValue: 'registry.local', description: 'Registry for base images')
        string(name: 'TARGET_REGISTRY', defaultValue: 'registry.local:5000', description: 'Registry for output images')
        string(name: 'SOURCE_CRED_ID', defaultValue: 'source-registry-creds')
        string(name: 'TARGET_CRED_ID', defaultValue: 'target-registry-creds')
    }
    environment {
        APP_NAME = "coin-api"
    }
    stages {
        stage('Test') {
            steps {
                script {
                    // Пример шага тестирования (при необходимости)
                    echo "Running tests for ${env.APP_NAME}..."
                    // sh 'go test ./...' 
                }
            }
        }
        stage('Build & Push') {
            steps {
                script {
                    // Настройка auth для нескольких registries (пример)
                    withCredentials([
                        usernamePassword(credentialsId: params.SOURCE_CRED_ID, passwordVariable: 'SRC_PASS', usernameVariable: 'SRC_USER'),
                        usernamePassword(credentialsId: params.TARGET_CRED_ID, passwordVariable: 'TGT_PASS', usernameVariable: 'TGT_USER')
                    ]) {
                        // Формирование docker config.json средствами Groovy, а не bash
                        def dockerConfig = """
                        {
                            "auths": {
                                "${params.SOURCE_REGISTRY}": { "auth": "${"${SRC_USER}:${SRC_PASS}".bytes.encodeBase64().toString()}" },
                                "${params.TARGET_REGISTRY}": { "auth": "${"${TGT_USER}:${TGT_PASS}".bytes.encodeBase64().toString()}" }
                            }
                        }
                        """
                        writeFile file: '/tmp/config.json', text: dockerConfig
                        
                        def targetImage = "${params.TARGET_REGISTRY}/${env.APP_NAME}:${env.BUILD_ID}"
                        
                        // Сборка через buildctl
                        sh """
                            export DOCKER_CONFIG=/tmp
                            buildctl build \\
                                --frontend dockerfile.v0 \\
                                --local context=./${env.APP_NAME} \\
                                --local dockerfile=./${env.APP_NAME} \\
                                --opt build-arg:REGISTRY=${params.SOURCE_REGISTRY}/ \\
                                --output type=image,name=${targetImage},push=true
                        """
                    }
                }
            }
        }
    }
}
```

## Как использовать скилл

1. Если пользователь просит "создать пайплайн" или "написать CI/CD", проверь целевой модуль.
2. Проверь наличие `Dockerfile` — если он требует доработок, исправь его, ориентируясь на `.cursor/skills/dockerfile/SKILL.md` и переносимость базовых образов (ARG).
3. Создай `Jenkinsfile` в корне модуля (или корне проекта, в зависимости от структуры), строго используя Groovy и `buildctl`.
4. Напиши `helm-chart-requirements.md` рядом с `Jenkinsfile` для фиксации требований к деплою в кластер.
5. Создай `corp-migration.md` рядом с пайплайном с конкретными инструкциями по адаптации модуля под корпоративную среду.
