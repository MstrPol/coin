# Troubleshoot CI (Control Plane v2)

**Ticket:** P1-07  
**Аудитория:** команды продуктов, platform on-call.

## Быстрая диагностика

| Этап | Что проверить |
|------|----------------|
| Resolve | `.coin/config.yaml`, coin-api `/ready`, Nexus cache |
| Bootstrap | podman system service; buildpack: paketo load |
| Pod | K8s cloud, `make endpoints`, `manifest.runtime.image` (coin-agent) |
| Stages | `coin-executor run --stage`, `manifest.build.engine` |
| Report | `coin-api-token`, `POST /v1/builds/report` |

## Resolve manifest

### 404 Golden path / version not found

**Симптом:** `curl …/manifest` → 404 или Jenkins: `curl: (22) The requested URL returned error: 404`.

**Причины:**

- Опечатка в `coin.goldenPath` или `coin.version`
- GP ещё не опубликован в catalog (Wave 1 — только `go-app@1.0.0`)

**Действия:**

```bash
curl -sf http://coin-api:8090/v1/golden-paths/go-app/versions/1.0.0/manifest | jq '.goldenPath'
# или с auth:
curl -sf -H "Authorization: Bearer $COIN_API_TOKEN" \
  http://coin-api:8090/v1/golden-paths/go-app/versions/1.0.0/manifest
```

Исправить `.coin/config.yaml` или дождаться publish GP.

### 401 Unauthorized

**Симптом:** Resolve или Report падает с HTTP 401.

**Причины:** `AUTH_DISABLED=false` в prod, но Jenkins credential `coin-api-token` отсутствует или не совпадает с `COIN_API_TOKEN` coin-api.

**Действия:**

1. Jenkins → Manage Credentials → `coin-api-token` (Secret text)
2. Значение = `COIN_API_TOKEN` на coin-api
3. Пересобрать job

Локально в `.env`: `AUTH_DISABLED=true` — token опционален для curl, но Jenkinsfile всё равно использует credential.

### Bearer: not found (shell quoting)

**Симптом:** `Bearer: not found` в console log Resolve/Report.

**Причина:** неправильное quoting переменной `AUTH=-H Authorization: Bearer …` в shell.

**Fix:** явный заголовок в curl:

```bash
curl -fsS -H "Authorization: Bearer ${COIN_API_TOKEN}" …
```

См. актуальный [`Jenkinsfile.coin`](../../coin-starters/Jenkinsfile.coin).

### coin-api unavailable — fallback Nexus cache

**Симптом:** в логе `coin-api unavailable — fallback Nexus cache`, build продолжается.

**Ожидаемо:** CI resilient design. Если fallback тоже падает — см. [api-down-nexus-fallback.md](../runbooks/api-down-nexus-fallback.md).

## K8s agent / pod

### Agent offline, JNLP Connection refused

**Симптом:** pod pending, `offline`, `Connect timed out` к Jenkins.

**Причина:** после restart Docker/k3s Endpoints `jenkins` указывают на неверный IP.

**Fix (local):**

```bash
cd docker && make endpoints
```

### lookup nexus: no such host (docker build)

**Симптом:** `docker push` / build не резолвит registry из pod.

**Fix:** `COIN_REGISTRY_PREFIX=localhost:8082/coin-docker` (host docker.sock). Уже в `Jenkinsfile.coin`.

### exec format error (coin-executor)

**Симптом:** `./coin-executor: cannot execute binary file` или `Exec format error`.

**Причина:** binary arch не совпадает с k3s node (arm64 vs amd64).

**Fix:** опубликовать оба:

```
maven-releases/coin/executor/coin-executor/0.1.0/coin-executor-0.1.0-linux-arm64
maven-releases/coin/executor/coin-executor/0.1.0/coin-executor-0.1.0-linux-amd64
```

Manifest `executor` секции нет — binary baked в `coin-agent`. Проверьте arch бинарника в образе agent.

### coin-executor job не публикует `coin-agent`

**Симптом:** job `coin-executor` падает на publish стадии до регистрации draft.

**Проверки:**

1. Jenkins credentials:
   - `nexus-docker` (user/password для `localhost:8082`)
   - `coin-publisher-api-key` (API key для `POST /v1/admin/components/agent/...`)
2. Docker доступен в Jenkins runner:
   - `docker version` в логе preflight
3. Доступность Nexus/coin-api из Jenkins контейнера:
   - `http://nexus:8082`
   - `http://coin-api:8090/ready`

**Ожидаемый результат после успешного run:**

- в Nexus Docker есть `coin-agent:<version>`
- в Platform появляется `agent/coin-agent@<version>` со статусом `draft`
- promote выполняется вручную в UI (job не делает auto-promote)

## coin-executor stages

### manifest sha256 mismatch / validate failed

**Симптом:** validate падает на hash content (`containerfile`, `validateSchema`).

**Причина:** Nexus immutable — metadata gp-content обновили без нового blob/version.

**Действия:**

```bash
# Новый gp-content semver + GP release (не перезапись blob)
cd docker && make seed-jenkins-lib
# или вручную: publish gp-content, затем POST GP version с новым pin
```

### Pod TerminationByKubelet (ephemeral-storage)

**Симптом:** buildpack job pending/killed, events: `ephemeral-storage: available: 0`.

**Причина:** k3s disk full (coin-agent ~3GiB + podman graph + paketo load).

**Fix:**

```bash
cd docker && bash scripts/prune-k3s-disk.sh --all
make e2e-build-engines   # включает prune перед прогоном
```

### podman short-name did not resolve

**Симптом:** `golang:1.25` / `gcr.io/...` не резолвятся в build.

**Fix:** в `coin-agent` image `podman-registries.conf` с `unqualified-search-registries = ["docker.io"]`. Пересобрать: `make publish-agent`.

### buildctl RUN /bin/sh invalid argument (arm64)

**Симптом:** buildkit stage падает на RUN в Dockerfile.

**Ожидаемо на local pilot arm64:** coin-executor использует **podman build** fallback. Если в логе всё ещё buildctl — обновить coin-executor + coin-agent.

### Jenkins lib cache (старый bootstrap)

**Симптом:** в логе `buildkitd` или старый Groovy после `make coin-lib`.

**Fix:** `make coin-lib` очищает `caches/git-*` в Jenkins volume.

### executor 404 при bootstrap (superseded)

**Симптом:** устаревший flow — `coin-executor bootstrap` и `manifest.executor` удалены; binary baked в `coin-agent`.

**Fix:** `make publish-agent`, проверить `manifest.runtime.image` указывает на актуальный `coin-agent` tag.

## Report stage

### build report failed / не пишется в DB

**Симптом:** stage Report красный или `✓ build report sent` но нет row в `build_reports`.

**Проверки:**

```bash
# API жив
curl -sf http://coin-api:8090/ready

# ручной POST (dev)
curl -X POST -H "Authorization: Bearer $COIN_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"project":{"name":"test","groupId":"g"},"buildUrl":"http://test","result":"success"}' \
  http://coin-api:8090/v1/builds/report

# DB
docker compose exec postgres psql -U coin -d coin \
  -c "SELECT * FROM build_reports ORDER BY id DESC LIMIT 3;"
```

**Частые причины:** неверный token, `project.name` не совпадает с config, coin-api недоступен из pod (Report **не** имеет Nexus fallback — только best-effort).

## Jenkinsfile / checkout

### unstash failed / manifest missing

**Симптом:** `No such saved stash 'coin-manifest'`.

**Причина:** `checkout scm` после `unstash` или stash не создан на Resolve.

**Fix:** порядок в Jenkinsfile: Resolve (checkout → curl → stash) → unstash на agent node. См. эталон `Jenkinsfile.coin`.

## Мониторинг (platform)

```bash
curl -sf http://coin-api:8090/metrics | grep coin_resolve_duration
```

p99 resolve > 100ms на cache hit — проверить Postgres, Nexus latency.

## Эскалация

| Severity | Критерий | Runbook |
|----------|----------|---------|
| S1 | coin-api down > 15 min, fallback OK | [api-down-nexus-fallback.md](../runbooks/api-down-nexus-fallback.md) |
| S2 | fallback 404 для production GP | Platform — re-publish manifest cache |
| S3 | mass agent offline | `make endpoints`, k3s/Jenkins |

## Ссылки

- [local-dev-control-plane.md](local-dev-control-plane.md)
- [agent-build-model.md](../agent-build-model.md)
- [jenkins-setup.md](../jenkins-setup.md)
- [wave-migration-checklist.md](wave-migration-checklist.md)
