# Golden paths

Golden path — **готовый профиль доставки** от кода до реестра.  
Разработчик выбирает один шаблон (`coin.template`), платформа берёт на себя всё остальное: toolchain, тип сборки, артефакты, публикацию, агент.

---

## Что такое golden path

Это не «стек сборки» и не «тип артефакта» по отдельности, а **закрытый контракт**:

| Что задаёт шаблон | Пример |
|-------------------|--------|
| Toolchain | Java 17 + Gradle |
| Роль проекта | приложение / библиотека |
| Что собираем | OCI image, JAR, wheel |
| Куда публикуем | Docker registry, Maven, PyPI |
| Стадии pipeline | test → build → publish |
| Образ агента | `coin/ci-jvm-gradle:17` (см. `coin-jenkins-agents/`) |
| Managed Dockerfile | runtime-only шаблон из Coin (COPY артефактов после native build) |
| Когда публикуем | по умолчанию `when: tag` |

**Правило:** если поведение можно вывести из шаблона — оно **не** настраивается в `.coin/config.yaml` проекта.

---

## Именование

```
{stack}-{role}
```

| Часть | Значение | Примеры |
|-------|----------|---------|
| `stack` | Язык + сборщик | `java-gradle`, `java-maven`, `python-uv`, `go` |
| `role` | Роль артефакта | `app` — деплоится; `lib` — переиспользуется другими проектами |

Примеры:

- `java-gradle-app` — Java-приложение, Gradle, образ для деплоя
- `java-gradle-lib` — Java-библиотека, Gradle, JAR в Maven
- `go-app` — Go-сервис, образ для деплоя

---

## Матрица golden paths (v1)

### Приложения (`*-app`)

Деплоятся как сервис. Сборка → OCI image → Docker registry.

| Шаблон | Toolchain | Артефакт | Publish | Статус |
|--------|-----------|----------|---------|--------|
| `go-app` | Go | OCI image | Docker registry | ✅ v1 |
| `java-gradle-app` | Java 17 + Gradle | OCI image | Docker registry | ✅ v1 |
| `java-maven-app` | Java 17 + Maven | OCI image | Docker registry | ✅ v1 |
| `python-uv-app` | Python + uv | OCI image | Docker registry | ✅ v1 |
| `python-pip-app` | Python + pip | OCI image | Docker registry | ✅ v1 |

> **Дистрибутив для ПСИ (zip):** если организации нужен zip в Nexus помимо образа — это часть профиля `*-app`, а не отдельный шаблон. Планируется для `java-*-app` (roadmap).

### Библиотеки (`*-lib`)

Переиспользуются другими проектами. Без контейнеризации.

| Шаблон | Toolchain | Артефакт | Publish | Статус |
|--------|-----------|----------|---------|--------|
| `java-gradle-lib` | Java 17 + Gradle | JAR | Maven (Nexus) | 📋 запланировано |
| `java-maven-lib` | Java 17 + Maven | JAR | Maven (Nexus) | 📋 запланировано |
| `python-uv-lib` | Python + uv | wheel | PyPI / Nexus | 📋 запланировано |

### Node.js

| Шаблон | Toolchain | Артефакт | Publish | Статус |
|--------|-----------|----------|---------|--------|
| `node-app` | Node.js | OCI image | Docker registry | 📋 запланировано |

---

## Что задаёт проект

Минимальный `.coin/config.yaml` — только привязка к GP, credentials и идентичность:

```yaml
coin:
  template: java-gradle-app
  templateVersion: v1

jenkins:
  credentials:
    docker: nexus-docker

project:
  name: my-service
  groupId: com.example.team
  repository: Nexus_PROD
```

### Что **не** задаётся в проекте

| Поле | Почему не нужно |
|------|-----------------|
| `build.type` | Определяется шаблоном (`app` → container, `lib` → package) |
| `agent.stack` | Дублирует `coin.template` |
| `container.port` / `container.command` | Задаётся в `profile.yaml` golden path |
| `dockerfileTemplate` | Зашито в profile шаблона |
| `publish.repository` (отдельно) | Тип publish задаёт шаблон; `project.repository` — координата для RN/QGM |

---

## Где живёт profile шаблона

Profile — platform-owned, разработчик не редактирует:

```
coin-golden-paths/
  catalog.yaml
  _shared/
    pack-image.sh
  java-gradle-app/
    v1/
      profile.yaml          # build, publish, pipeline, agent defaults
      Dockerfile            # runtime-only (COPY артефактов)
      scripts/
      config.yaml
```

Пример `profile.yaml`:

```yaml
agent:
  stack: java-gradle
  runtime:
    java: "17"

build:
  type: container
  dockerfile: Dockerfile

publish:
  kind: registry
  when: tag

pipeline:
  test:
    enabled: true
  build:
    enabled: true
  publish:
    enabled: true

container:
  port: 8080
  command: ["java", "-jar", "/app/app.jar"]
```

Coin CLI при `coin run build` загружает bundle по `coin.template` + `templateVersion` и выполняет сценарий.  
Модель сборки: **native compile в agent → runtime-only Dockerfile → registry**. Подробнее — [agent-build-model.md](agent-build-model.md), [golden-path-versioning.md](golden-path-versioning.md).

---

## Скелетоны новых проектов (`coin-starters/`)

Golden path **не копируется** в репозиторий сервиса целиком — там только platform-owned артефакты доставки.

Для bootstrap нового репозитория используйте **starter** — минимальный рабочий проект:

```
coin-starters/
  python-uv-app/
    .coin/config.yaml
    Jenkinsfile
    pyproject.toml
    src/
    tests/
  go-app/
    ...
```

```bash
cp -r coin-starters/python-uv-app/* /path/to/my-new-service/
# или
coin init
```

| Каталог | Владелец | Куда попадает |
|---------|----------|---------------|
| `coin-golden-paths/` | Platform | Загружается coin CLI в CI |
| `coin-starters/` | Platform (эталон) | Копируется командой в свой репо |

Подробнее — [coin-starters/README.md](../coin-starters/README.md).

---

## Разделение «repository»

В конфиге одно поле `project.repository`, но смысл зависит от контекста:

| Контекст | Что означает `project.repository` |
|----------|-----------------------------------|
| Release notes (QGM) | Логическое имя репозитория Nexus (`Nexus_PROD`) |
| Maven coordinates | Репозиторий для метаданных артефакта |
| Физический URL registry | **Не** в проекте — mapping платформы + credentials |

---

## Текущее состояние

1. Матрица `*-app` зафиксирована ✅
2. Структура `coin-golden-paths/<name>/v1/` + `catalog.yaml` ✅
3. `profile.yaml` в каждом v1 ✅
4. Упрощённый `.coin/config.yaml` в шаблонах ✅
5. `coin-starters/` — скелетоны для bootstrap ✅
6. `*-lib` шаблоны — по мере появления кейсов

Новые сервисы: `cp -r coin-starters/<name>/*` → свой репозиторий.

---

## Как выбрать шаблон

```
Это библиотека для других проектов?
  ├─ Да  →  {stack}-lib
  └─ Нет →  {stack}-app

Какой стек?
  ├─ Go                          →  go-app
  ├─ Java + Gradle               →  java-gradle-app / java-gradle-lib
  ├─ Java + Maven                →  java-maven-app / java-maven-lib
  ├─ Python + uv                 →  python-uv-app / python-uv-lib
  └─ Python + pip                 →  python-pip-app
```

---

## Связанные документы

- [agent-build-model.md](agent-build-model.md) — native build + runtime-only Dockerfile
- [config.md](config.md) — структура `.coin/config.yaml`
- [golden-path-versioning.md](golden-path-versioning.md) — v1/v2, доставка каталога
- [coin-starters/README.md](../coin-starters/README.md) — скелетоны новых репозиториев
- [responsibilities.md](responsibilities.md) — кто управляет шаблонами и конфигом
- [architecture.md](architecture.md) — как Coin CLI выполняет стадии pipeline
- [jenkins-setup.md](jenkins-setup.md) — platform и service jobs
