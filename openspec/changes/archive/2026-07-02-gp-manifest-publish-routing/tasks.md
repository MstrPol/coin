## 1. Архитектурный контракт

- [x] 1.1 Обновить ADR/docs: product config = GP pin + project identity + Jenkins glue; GP/Build Stack = build/publish policy; executor = исполнение
- [x] 1.2 Задокументировать `destinations` в GP manifest и fallback через Nexus
- [x] 1.3 Задокументировать, что credentials остаются вне manifest

## 2. Product config v2

- [x] 2.1 Удалить `project.repository` из schema, parser/validation, starters, samples и docs
- [x] 2.2 Запретить секцию `deliverables` в product `.coin/config.yaml`
- [x] 2.3 Добавить tests для тонкого config и отклонения repository/deliverables/build-publish fields

## 3. GP/Build Stack contract

- [x] 3.1 Добавить в GP version/release поля `imageRegistryPrefix`, `buildCacheEnabled`, `artifactRepositoryBase`
- [x] 3.2 Материализовать `destinations` в resolved manifest и включить их в `manifestHash`
- [x] 3.3 Перенести source of truth для deliverables в GP/Build Stack manifest materialization
- [x] 3.4 Ограничить P0 deliverables: максимум один `image`, один `liquibase-image`, один `artifact`
- [x] 3.5 Удалить `cacheRefTemplate` из gp-content schema/editor/preview/seed data

## 4. coin-executor

- [x] 4.1 Добавить `destinations` и GP/Build Stack deliverables в manifest structs
- [x] 4.2 Строить app image, liquibase image, cache ref и artifact repository URL из manifest destinations + project identity
- [x] 4.3 Брать список outputs из manifest, а не из product config
- [x] 4.4 Убрать source of truth из `COIN_REGISTRY_PREFIX`, `project.repository` и `build.*.cacheRef`
- [x] 4.5 Добавить tests для refs, cache on/off, artifact repository URL и запрета product deliverables

## 5. coin-ui / coin-api

- [x] 5.1 Обновить OpenAPI/schema для `destinations` и тонкого product config
- [x] 5.2 Обновить GP draft/release UI: поля destinations и read-only detail
- [x] 5.3 Обновить Build Stack editor: deliverables P0 и отсутствие `cacheRefTemplate`
- [x] 5.4 Обновить gp-content preview: без physical destinations/cache refs

## 6. coin-lib

- [x] 6.1 Материализовать manifest с `destinations` без изменений из API/Nexus fallback
- [x] 6.2 Оставить только credentials binding, pod и stage orchestration
- [x] 6.3 Убрать registry/cache defaults как source of truth для publish destinations

## 7. Local pilot validation

- [ ] 7.1 Пересеять local GP `gp-01-07@1.0.0` с destinations и GP-owned deliverables
- [x] 7.2 Обновить `samples/demo-go-app/.coin/config.yaml` до тонкого product config
- [ ] 7.3 Проверить resolved manifest: есть `destinations`, есть GP deliverables, нет `build.*.cacheRef`, нет credentials
- [ ] 7.4 Запустить focused tests для `coin-api`, `coin-executor`, `coin-ui`, `coin-lib`
- [ ] 7.5 Запустить E2E `samples/demo-go-app`: resolve → validate → test → build → publish
- [x] 7.6 Запустить `openspec validate gp-manifest-publish-routing --strict`
