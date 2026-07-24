# Java + Maven app (Coin starter)

Минимальный Java-сервис. Golden path: `java-maven-app`.

```bash
cp -r coin-starters/java-maven-app/* /path/to/my-service/
# или: coin init --starter java-maven-app
```

CI: `mvn test` → native JAR → pack runtime-only Dockerfile (JRE). См. [docs/agent-build-model.md](../../docs/agent-build-model.md).
