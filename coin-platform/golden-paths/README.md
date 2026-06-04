# coin-golden-paths

Каталог **golden paths** — platform-owned профили доставки от кода до реестра.

Каждый golden path — это не «шаблон файлов», а **закрытый контракт**: toolchain, тип сборки, скрипты CI, managed Dockerfile, политика publish.

```
coin-golden-paths/
  catalog.yaml                 # policy: latest, minimum per path
  _shared/
    pack-image.sh              # docker/kaniko: упаковка артефактов в OCI image
  python-uv-app/
    v1/
      profile.yaml             # build, publish, pipeline defaults
      Dockerfile               # runtime-only (не копируется в репо сервиса)
      scripts/                 # test.sh, build.sh, publish.sh
      config.yaml              # эталон .coin/config.yaml для нового сервиса
```

Сборка `*-app`: native compile в agent → `pack-image.sh` → registry. См. [docs/agent-build-model.md](../docs/agent-build-model.md).

## Выбор в проекте

```yaml
coin:
  template: python-uv-app
  templateVersion: v1
```

Матрица путей — [docs/golden-paths.md](../docs/golden-paths.md).  
Версионирование каталога — [docs/golden-path-versioning.md](../docs/golden-path-versioning.md).

## Загрузка в coin CLI

| Переменная | Описание |
|------------|----------|
| `COIN_GOLDEN_PATHS_SOURCE` | `local` (default) или `nexus` |
| `COIN_GOLDEN_PATHS_DIR` | явный путь к каталогу |
| `COIN_GOLDEN_PATHS_URL` | tarball из Nexus (для `nexus`) |

## Связанные документы

- [docs/golden-paths.md](../docs/golden-paths.md)
- [docs/agent-build-model.md](../docs/agent-build-model.md)
- [docs/golden-path-versioning.md](../docs/golden-path-versioning.md)
