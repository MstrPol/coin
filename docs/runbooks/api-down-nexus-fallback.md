# Runbook: coin-api недоступен — Nexus manifest fallback

**Ticket:** PF-03 / P1-07  
**Severity:** S1 (CI продолжает работать при cache hit) / S2 (cache miss или stale).

## Контекст

Resolve manifest в Jenkins:

1. **Primary:** `GET /v1/golden-paths/{gp}/versions/{ver}/manifest` (coin-api)
2. **Fallback (exact pin):** Nexus pointer → immutable blob → verify `manifestHash`

```
maven-snapshots/coin/manifest/{gp}/metadata/{gp}-metadata-pin-%3D{version}.json  →  pointer
maven-releases/coin/manifest/{gp}/{version}/{gp}-{version}.json                  →  blob (immutable)
```

Report stage **не** имеет fallback — только POST в coin-api (best-effort; build не fail по дизайну executor).

## Симптомы

| Наблюдение | Интерпретация |
|------------|---------------|
| Log: `coin-api unavailable — fallback Nexus pointer → blob` + SUCCESS | **OK** — resilient path |
| Resolve FAIL после fallback | Pointer/blob отсутствует или GP/version новый |
| Mass builds FAIL на Resolve | API down + Nexus cache проблема |
| Report warnings | API down — builds green, analytics gap |

## Диагностика (< 5 min)

```bash
GP=go-app
VER=1.0.0
SNAPSHOTS=http://localhost:8081/repository/maven-snapshots

# 1. API health
curl -sf http://coin-api:8090/ready || echo "API DOWN"

PTR_PATH="${SNAPSHOTS}/coin/manifest/${GP}/metadata/${GP}-metadata-pin-%3D${VER}.json"

# 2. Pointer exists (exact pin =1.0.0 → %3D1.0.0)
curl -sf -o /dev/null -w "%{http_code}\n" "${PTR_PATH}"

# 3. Pointer → blob hash match
PTR=$(curl -sf "${PTR_PATH}")
BLOB=$(echo "$PTR" | jq -r .blobUrl)
EXPECTED=$(echo "$PTR" | jq -r .manifestHash)
ACTUAL=$(curl -sf "$BLOB" | jq -r .manifestHash)
echo "expected=$EXPECTED actual=$ACTUAL"

# 4. coin-api logs
docker compose -f docker/compose.yml logs --tail=100 coin-api
```

Jenkins console (Resolve stage): первая curl к API, при fail — pointer → blob + `test manifestHash`.

## Действия по сценарию

### A. API down, Nexus pointer OK (S1)

**Impact:** builds продолжают работать. Report может не писаться.

1. On-call platform: восстановить coin-api / Postgres
2. Коммуникация: «CI green, fleet analytics delayed»
3. После восстановления API: проверить `/ready`, metrics

```bash
cd docker
docker compose up -d coin-api postgres
curl -sf http://localhost:8090/ready
```

### B. API down, pointer 404 (S2)

**Impact:** новые/обновлённые GP versions не резолвятся.

1. Проверить path: `maven-snapshots/coin/manifest/{gp}/metadata/{gp}-metadata-pin-%3D{version}.json`
2. Прогреть cache через resolve при живом API:

```bash
curl -fsS http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest >/dev/null
```

3. Или E2E script:

```bash
cd docker && make e2e-mvp1
```

4. Восстановить API для canonical resolve + auto-upload Nexus

### C. manifestHash mismatch после восстановления API

**Impact:** validate fail на executor / Resolve stage.

1. Re-resolve manifest (coin-api перезаливает blob + pointer)
2. Rebuild coin-api если изменился expected hash в store
3. Перезапустить failed builds

### D. Planned maintenance API

**До окна:**

- [ ] Убедиться что все production GP versions имеют pointer + blob в Nexus
- [ ] `make e2e-mvp1` green
- [ ] Comms: builds на Nexus fallback, report gap acceptable

**После окна:**

- [ ] `/ready` green
- [ ] Spot-check 3 repos: Resolve без fallback log line
- [ ] Report stage на canary build

## Профилактика

| Мера | Когда |
|------|-------|
| Auto-upload blob+pointer+content при Resolve | coin-api (MVP-1) |
| Мониторинг `/ready` + alert | P1-02 metrics |
| HA coin-api ≥2 replicas | **P1-05 deferred** — post Wave 50 |
| Pre-publish cache при GP release | P2 publish path |

## Связанные документы

- [troubleshoot-ci.md](../how-to/troubleshoot-ci.md)
- [control-plane.md](../control-plane.md) — три слоя SoT
- [local-dev-control-plane.md](../how-to/local-dev-control-plane.md)
