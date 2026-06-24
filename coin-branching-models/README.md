# coin-branching-models

Каталог reference-моделей ветвления для GP composition slot `branching-model`.

## Layout

```
models/
├── trunk-based/     # сервисы: go-app, go-app-bp, go-app-df
│   ├── model.yaml
│   └── README.md
└── semver-tag/      # библиотеки: go-lib, java-maven-app (future)
    ├── model.yaml
    └── README.md
schemas/
└── branching-model.schema.json
scripts/
├── publish-branching-model.sh
└── lib/maven-url.sh
```

Primary artifact: `model.yaml`. Resolve materializer читает `content_ref` v2 с manifest subset `branching`.

## Publish

**Primary path:** Component Studio (`/studio` или `/branching-models`) → validate → register (PG) → canary → promote (Nexus).

**Bootstrap / local seed** (deprecated для fleet, OK для pilot):

```bash
./scripts/publish-branching-model.sh trunk-based 1.0.0
./scripts/publish-branching-model.sh semver-tag 1.0.0
```

Flow: draft → artifact `model.yaml` → `register-package` (PG content_ref v2, без Nexus) → canary → promote (Nexus + published).

```bash
cd docker
make seed-jenkins-lib   # включает trunk-based@1.0.0 в 5-slot composition
```

## См. также

- [docs/adr/gp-branching-model.md](../docs/adr/gp-branching-model.md)
- [docs/branching.md](../docs/branching.md)
- [docs/golden-paths.md](../docs/golden-paths.md)
