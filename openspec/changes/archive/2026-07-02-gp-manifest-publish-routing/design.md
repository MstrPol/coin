## Контекст

Цель Coin — убрать build/publish логику из продуктовых репозиториев. Продукт выбирает GP и передаёт identity, а платформа через GP/Build Stack определяет, что собрать, как собрать и куда опубликовать.

Build path должен работать без `coin-api`, поэтому всё, что нужно для сборки и публикации, должно быть в resolved manifest, который кешируется в Nexus.

## Цели / Не цели

**Цели:**

- Сделать `.coin/config.yaml` тонким: GP pin, project identity и project-specific Jenkins glue.
- Перенести deliverables, build policy и publish destinations под контроль GP/Build Stack.
- Сделать manifest достаточным для fallback build path.
- Оставить `coin-lib` glue-only.
- Оставить Jenkins credential IDs вне manifest.

**Не цели:**

- Не давать разработчикам писать build/publish commands в product repo.
- Не давать product repo выбирать или расширять deliverables.
- Не вводить отдельный destination service/catalog/artifact.
- Не проектировать corp multi-environment destination model.
- Не поддерживать несколько outputs одного типа в P0.
- Не добавлять PyPI/npm/другие package ecosystems.

## Решения

### Решение 1: product config остаётся тонким

Product `.coin/config.yaml` содержит только:

```yaml
coin:
  goldenPath: gp-01-07
  version: "1.0.0"

project:
  name: demo-go-app
  groupId: com.example.team
  artifactId: demo-go-app

jenkins:
  credentials:
    docker: nexus-docker
```

Запрещены `project.repository`, `deliverables`, build/publish commands, pipeline stages, cache refs и registry/repository URLs.

### Решение 2: GP/Build Stack задаёт outputs

Deliverables являются частью GP/Build Stack policy и материализуются в manifest. Для P0 разрешён максимум один output каждого типа:

```yaml
deliverables:
  app:
    type: image
  liquibase:
    type: liquibase-image
  artifact:
    type: artifact
```

Проект получает весь набор outputs выбранного GP. Если нужен другой набор, platform team выпускает другую версию GP или отдельный GP.

### Решение 3: GP version задаёт destinations

Resolved manifest содержит:

```json
{
  "destinations": {
    "imageRegistryPrefix": "docker-dev.registry.domain.ru",
    "buildCacheEnabled": true,
    "artifactRepositoryBase": "http://nexus:8081/repository/maven-releases"
  }
}
```

`imageRegistryPrefix` включает registry host и repository/namespace. Отдельный `cacheRegistryPrefix` не вводится: cache ref строится от `imageRegistryPrefix`. `artifactRepositoryBase` задаёт Nexus repository URL для artifact publish.

### Решение 4: executor строит refs

```text
path = project.groupId + "/" + project.artifactId + "/" + project.name

appImageRef =
  destinations.imageRegistryPrefix + "/" + path + ":" + resolvedVersion

liquibaseImageRef =
  destinations.imageRegistryPrefix + "/" + project.groupId + "/" +
  project.artifactId + "/" + project.name + "-liquibase:" + resolvedVersion

cacheRef =
  destinations.imageRegistryPrefix + "/" + path + "-cache

artifactRepoURL =
  destinations.artifactRepositoryBase
```

Cache используется только при `buildCacheEnabled: true`.

### Решение 5: `gp-content` не хранит physical destinations

`gp-content` описывает engine, targets, Containerfile, pipeline и deliverables. Он не хранит `cacheRefTemplate`, registry URLs, repository URLs или credentials.

### Решение 6: credentials остаются вне manifest

Manifest может содержать physical URL/prefix в `destinations`, но не Jenkins credential IDs. Credentials остаются в product/Jenkins glue layer.

## Риски / Trade-offs

- **GP становится более жёстким** → это целевое поведение hard Golden Path.
- **Product repo не может выбрать subset outputs** → другой набор outputs оформляется новой GP version или отдельным GP.
- **GP release становится environment-specific** → для pilot принимаем; corp destination model вынести отдельным change после появления реальной потребности.
- **Один output каждого типа может быть тесным** → multi-image/multi-artifact вынести отдельным change.

## План миграции

1. Добавить `destinations` к существующей GP version/release.
2. Материализовать `destinations` и GP/Build Stack deliverables в manifest.
3. Удалить `project.repository` и `deliverables` из product config v2.
4. Удалить `cacheRefTemplate` из gp-content schema/editor/preview/seed data.
5. Научить executor брать outputs из manifest и строить refs из `destinations` + project identity.
6. Оставить `coin-lib` только glue: resolve, fallback, pod, credentials binding, stages.
7. Обновить starters/samples/docs.
8. Пересеять local GP и прогнать E2E `samples/demo-go-app`.

## Открытые вопросы

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Нужен ли отдельный destination layer после corp gate? | ✅ decided for pilot | A: поля в GP version; B: отдельный catalog/artifact | A на pilot |
