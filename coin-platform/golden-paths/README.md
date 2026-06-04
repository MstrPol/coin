# Golden paths

Профили доставки от кода до реестра — часть [coin-platform](../README.md).

```
golden-paths/
  catalog.yaml
  _shared/pack-image.sh
  python-uv-app/v1/
    profile.yaml
    Dockerfile          # runtime-only
    scripts/
    config.yaml         # эталон .coin/config.yaml
```

Загрузка в coin CLI: `$COIN_PLATFORM_DIR/golden-paths` или `COIN_GOLDEN_PATHS_DIR`.

Документация: [docs/golden-paths.md](../../docs/golden-paths.md).
