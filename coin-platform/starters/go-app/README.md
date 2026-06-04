# Go app (Coin starter)

Минимальный Go-сервис. Golden path: `go-app`.

```bash
cp -r coin-starters/go-app/* /path/to/my-service/
# или: coin init --starter go-app
```

CI: `go test` → `go build` (native) → pack runtime-only Dockerfile → registry.  
Dockerfile **не** в репо — managed Coin. См. [docs/agent-build-model.md](../../docs/agent-build-model.md).
