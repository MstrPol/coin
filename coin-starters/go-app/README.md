# Go app (Coin starter)

Минимальный Go-сервис. Golden path: `go-app`.

```bash
cp -r coin-starters/go-app/* /path/to/my-service/
cp coin-starters/Jenkinsfile.coin /path/to/my-service/Jenkinsfile
# или: make samples (локальный E2E → Gitea + Jenkins multibranch)
```

CI: `go test` → `go build` (native) → pack runtime-only Dockerfile → registry.  
Dockerfile **не** в репо — managed Coin. См. [docs/agent-build-model.md](../../docs/agent-build-model.md).
