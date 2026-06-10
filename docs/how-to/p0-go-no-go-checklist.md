# Go/no-go: фаза 0 (pilot)

**Роль:** Lead. **Gate:** перед Wave 1 (50 repos).

**Решение:** **GO** — 2026-06-05  
**Стенд:** local docker (`make bootstrap` + `make endpoints`)  
**Эталон:** `demo-go-app@main` build **#13 SUCCESS** (validate → test → build → report)

## Checklist

| # | Критерий | Статус | Verify |
|---|----------|--------|--------|
| 1 | demo-go-app E2E green без lib/cli | ✅ | Jenkins `demo-go-app/main` #13 SUCCESS + Report |
| 2 | Manifest в Nexus cache | ✅ | pointer `pointers/go-app/%3D1.0.0.json` + blob |
| 3 | coin-api `/ready` | ✅ | `curl localhost:8090/ready` → `{"status":"ready"}` |
| 4 | Resolve + fallback | ✅ | coin-api stopped → Nexus manifest readable |
| 5 | Hard cut lib/cli | ✅ | `go-app/Jenkinsfile` без `@Library`; README DEPRECATED |
| 6 | Docs P0-18 | ✅ | how-to×3 + `docs/README.md` v2 |
| 7 | `make endpoints` documented | ✅ | `docker/README.md`, local-dev how-to |

## Протокол

| Поле | Значение |
|------|----------|
| Дата | 2026-06-05 |
| Решение | **GO** → Wave 1 planning (P1-06) |
| Pilot GP | `go-app@1.0.0` |
| E2E scope | validate → test → build (без publish tag) |
| Блокеры | нет |

### Verify commands (повтор)

```bash
curl -sf http://localhost:8090/ready
BASE=http://localhost:8081/repository/coin-manifests
curl -sf "${BASE}/pointers/go-app/%3D1.0.0.json" | jq .manifestHash
curl -sf -u admin:admin http://localhost:8080/job/demo-go-app/job/main/lastBuild/api/json?tree=result
```

### Fallback test (2026-06-05)

```bash
docker compose stop coin-api
curl -sf "${BASE}/pointers/go-app/%3D1.0.0.json"  # OK
docker compose start coin-api
cd docker && make e2e-mvp1
```

## Известные ограничения pilot (accept)

- Только GP `go-app@1.0.0`
- coin-api single instance (HA — P1-05 deferred)
- Auth Bearer + Report — ✅ P1-01…P1-04 (build #13)
- java/python starters — ещё v1 Jenkinsfile (миграция Wave 2+)

## Следующий шаг

**Local dev:** P2-02 Admin publish GP, coin-ui (samples only).  
**Corp gate:** Wave rollout по [wave-migration-checklist.md](wave-migration-checklist.md).

## Ссылки

- [local-dev-control-plane.md](local-dev-control-plane.md)
- [ADR control-plane-v2](../../.cursor/plans/adr/control-plane-v2.md)
- [platform-native-jenkins.plan.md](../../.cursor/plans/platform-native-jenkins.plan.md)
