## Зачем

Coin выбирает жёсткий Golden Path: разработчик не пишет build/publish tasks в продукте, а выбирает версию GP, где платформа уже определила build stack, publish outputs и destinations. Сейчас часть этих решений всё ещё живёт вне GP: cache ref вычисляется из `gp-content`, registry prefix приходит из `coin-lib`, а product config хранит `project.repository`.

Для pilot переносим build/publish policy в GP/Build Stack и делаем resolved manifest самодостаточным для fallback через Nexus.

## Что меняется

- **BREAKING**: product `.coin/config.yaml` становится тонким: `coin.goldenPath`, `coin.version`, `project.name`, `project.groupId`, `project.artifactId`, project-specific Jenkins glue.
- **BREAKING**: из product config удаляются `project.repository`, `deliverables`, build/publish commands, pipeline stages, cache refs и registry/repository URLs.
- **BREAKING**: GP/Build Stack полностью задаёт P0 deliverables: максимум один `image`, максимум один `liquibase-image`, максимум один `artifact`.
- **BREAKING**: версия GP получает поля `destinations`: `imageRegistryPrefix`, `buildCacheEnabled`, `artifactRepositoryBase`.
- `gp-content` больше не хранит `cacheRefTemplate`.
- `coin-executor` строит image/cache/artifact refs из `manifest.destinations` и project identity.
- `coin-lib` остаётся Jenkins glue и не становится SoT для destinations.
- Resolved manifest по-прежнему не содержит Jenkins credential IDs.

## Capabilities

### New Capabilities

- `gp-manifest-publish-fields`: поля destinations в существующей версии GP и их материализация в resolved manifest.
- `deliverable-publish-contract`: P0-контракт outputs, полностью управляемый GP/Build Stack.
- `product-config-v2`: тонкий product config без repository, deliverables и build/publish logic.

### Modified Capabilities

- `gp-composition-two-slot`: manifest материализуется из GP identity, destinations версии GP и трёх component pins.
- `build-engine`: executor строит publish/cache refs из manifest destinations.
- `platform-build-stacks`: Build Stack больше не редактирует `cacheRefTemplate`.
- `gp-content-preview`: preview не возвращает cache refs и physical destination values.

## Влияние

- `coin-api`: GP version/release payloads, persistence fields, manifest builder/schema/OpenAPI, product config schema, GP/Build Stack deliverables validation.
- `coin-ui`: GP release form/detail для destinations и Build Stack editor без `cacheRefTemplate`.
- `coin-executor`: manifest structs, output selection из manifest, ref construction.
- `coin-lib`: materialize manifest and bind credentials only.
- `coin-gp-content`: stack content/schema без `cacheRefTemplate`, deliverables как часть GP/Build Stack policy.
- `coin-starters` / `samples`: тонкий `.coin/config.yaml`.
- Docs/ADR: зафиксировать pilot trade-off и границы product config / GP / executor.

## Не цели

- Не вводим отдельный destination service/catalog/artifact.
- Не вводим новый platform component type или отдельную destination model/table.
- Не храним destinations в Jenkins.
- Не добавляем Jenkins credential IDs в manifest.
- Не проектируем corp multi-environment destination model.
- Не поддерживаем PyPI/npm/другие package ecosystems.
- Не даём product repo расширять или выбирать subset deliverables.
- Не поддерживаем несколько deliverables одного типа в одном GP/Build Stack.
