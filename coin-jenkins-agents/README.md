# coin-jenkins-agents

CI agent images для Jenkins dynamic agents (K8s pod, контейнер `stack`).

**Corp target repo:** `coin/coin-jenkins-agents` (PF-16 split из monolithic `coin-platform`).

## Сборка

Job **`agents-build`** — `Jenkinsfile` в корне repo.

```bash
cd docker && make agents-build   # регистрация job
cd docker && make coin-jenkins-agents   # push → Gitea (local pilot)
```

## catalog.yaml

Manifest agent images. Job **пишет** `rev`, `tag`, `digest` после каждой сборки.

`make coin-jenkins-agents` перед push подтягивает `catalog.yaml` из Gitea, если локально не меняли с прошлого push.

Полный ref: `{registry.default}/{image}:{tag}` → `nexus:8082/coin-docker/ci-go:1.22-r1`.

## Связь с GP

Composition slot `agent` в GP release → `manifest.runtime.image` в pod template.

Legacy v1 `profile.yaml` (agent.stack/runtime) — superseded; см. GP composition в coin-api.

## Build context

Корень repo (Dockerfile paths в catalog относительно этой папки).
