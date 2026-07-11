# Миграция config v1 → v2

**Цель:** перевести продуктовый репозиторий с config v1 на Control Plane v2.

**Gate:** P0 go/no-go.

## Mapping полей

| v1 | v2 |
|----|-----|
| `coin.template: go-app` | `coin.goldenPath: go-app` |
| `coin.templateVersion: v1` | `coin.version: "1.0.0"` |
| `jenkins.runtime.*` | удалить — agent в manifest |
| `pipeline.*` overrides | удалить на pilot — stages в manifest |

`v1` каталог GP **≠** semver `"1.0.0"`. Для go-app pilot используйте release `1.0.0` из catalog.

## Шаги

### 1. Обновить `.coin/config.yaml`

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"

jenkins:
  credentials:
    docker: nexus-docker

project:
  name: my-service
  groupId: com.example.team
  artifactId: my-service
```

### 2. Заменить Jenkinsfile

Удалить v1 Jenkinsfile (Shared Library + `coinPipeline()`).

Скопировать [`coin-starters/Jenkinsfile.coin`](../../coin-starters/Jenkinsfile.coin).

### 3. Удалить из репо (если есть)

- `.coin/platform/` checkout
- `.coin/generated/Dockerfile`
- Локальные CI scripts, дублирующие GP

### 4. Verify локально (optional)

```bash
curl -fsS http://coin-api:8090/v1/golden-paths/go-app/versions/1.0.0/manifest -o .coin/manifest.json
coin-executor validate --project .coin/config.yaml --manifest .coin/manifest.json
```

### 5. Push + Jenkins build

Первый build на v2: resolve → pod → validate → test → build.

## Troubleshooting

| Ошибка | Причина |
|--------|---------|
| `coin.goldenPath is required` | Остались v1 поля `template` |
| `manifest gp != config` | Неверный `version` |
| Agent offline | `make endpoints` на стенде |

## Rollback

Hard cut — откат на config v1 **не поддерживается** на pilot. Исправлять forward.

## Ссылки

- [config.md](../config.md)
- [add-new-service-repo.md](add-new-service-repo.md)
