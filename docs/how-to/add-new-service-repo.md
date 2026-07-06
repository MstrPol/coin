# Новый сервисный репозиторий (go-app)

**Цель:** создать polyrepo с Control Plane v2 с нуля.

**Gate:** P0 go/no-go. Pilot GP: **go-app** (buildkit); также **go-app-bp**, **go-app-df** для E2E engines.

## Prerequisites

- Control Plane стенд или corp Jenkins + coin-api
- Gitea / GitHub repo для сервиса
- Jenkins multibranch на repo

## Шаги

### 1. Starter

```bash
cp -r coin-starters/go-app/ my-service/
cd my-service
```

Содержимое starter:

- `.coin/config.yaml` — v2
- `Jenkinsfile` — копия `Jenkinsfile.coin`
- минимальный Go код + тест

### 2. Настроить config

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"

jenkins:
  credentials:
    docker: nexus-docker   # Jenkins credential ID в вашем Jenkins

project:
  name: my-service
  groupId: com.example.team
  artifactId: my-service
```

### 3. Jenkinsfile

Должен совпадать с [`Jenkinsfile.coin`](../../coin-starters/Jenkinsfile.coin).  
Env overrides (optional):

- `COIN_API_URL` — default `http://coin-api:8090`
- `COIN_MANIFEST_CACHE_BASE` — Nexus manifests repo

### 4. Git push

```bash
git init -b main
git add .
git commit -m "feat: initial go-app service"
git remote add origin <your-repo-url>
git push -u origin main
```

### 5. Jenkins multibranch

Создать multibranch job → scan → build `main`.

## Verify

- [ ] Resolve manifest без ошибок
- [ ] Stages: Validate, Test, Build — SUCCESS (engine из GP manifest, не из проекта)
- [ ] Docker image в registry (local: `localhost:8082/coin-docker/app:<build>`)

```bash
# на стенде после build
curl -sf http://localhost:8082/v2/coin-docker/app/tags/list | jq .
```

## Troubleshooting

| Симптом | Действие |
|---------|----------|
| GP not found (404) | Проверить `goldenPath`/`version` в catalog API |
| Нет executor binary | Platform job `coin-executor` publish |
| Pod pending/offline | `make endpoints` (local k3s) |

## Ссылки

- [config.md](../config.md)
- [branching-models.md](branching-models.md) — GP-pinned branching model (`manifest.branching`)
- [local-dev-control-plane.md](local-dev-control-plane.md)
- [samples/demo-go-app/](../../samples/demo-go-app/) — эталон E2E
