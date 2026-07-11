# Java + Gradle app (Coin starter)

Минимальный Java-сервис. Golden path: `java-gradle-app`.

```bash
cp -r coin-starters/java-gradle-app/* /path/to/my-service/
# или: coin init --starter java-gradle-app
```

CI: `gradle test` → native JAR → pack runtime-only Dockerfile (JRE). См. [docs/agent-build-model.md](../../docs/agent-build-model.md).
