# Python + uv app (Coin starter)

Минимальный Python-сервис. Golden path: `python-uv-app`.

```bash
cp -r coin-starters/python-uv-app/* /path/to/my-service/
# или: coin init --starter python-uv-app
```

CI: `uv run pytest` → native deps → pack runtime-only Dockerfile → registry.  
Dockerfile **не** в репо — managed Coin. См. [docs/agent-build-model.md](../../docs/agent-build-model.md).
