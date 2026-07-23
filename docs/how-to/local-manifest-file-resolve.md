# Local manifest resolve (`coin.resolve: file`)

Для отладки **coin-executor** и формы resolved-манифеста без coin-api / Nexus.

## Config

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
  resolve: file
  # manifestFile: .coin/manifest.local.yaml   # default
```

`goldenPath` и `version` остаются обязательными (identity / report).

## Fixture

Положите resolved-shape документ в `.coin/manifest.local.yaml` (YAML или JSON) и файлы Containerfile в `.coin/containerfiles/<id>` (catalog указывает только `id` / `kind` / `path`, **без** inline `body`).

После resolve coin-lib materialize’ит `.coin/manifest.json` (gitignore) как при remote path.

В логе Jenkins будет soft-warn: local file resolve не для production product repos.

## См. также

- ADR: [pipeline-tekton-mapping.md](../adr/pipeline-tekton-mapping.md)
- Schema: [pipeline-inline.v4.schema.json](../schemas/pipeline-inline.v4.schema.json)
- Sample: `samples/demo-go-app` с `resolve: file` (`pipeline-tekton-alignment`)
- API/UI/remote E2E: change `pipeline-v4-control-plane`
