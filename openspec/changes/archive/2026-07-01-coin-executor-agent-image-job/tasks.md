## 1. Pipeline contract and prerequisites

- [x] 1.1 Уточнить параметры Jenkins job (`BUMP`, `GOARCH`, registry/api env) и зафиксировать ожидаемый output version/image tag
- [x] 1.2 Проверить и задокументировать обязательные Jenkins credentials (`nexus-docker`, `coin-publisher-api-key`) для publish/draft flow
- [x] 1.3 Убедиться, что node для job имеет доступ к Docker daemon и registry `localhost:8082`

## 2. Jenkinsfile rewrite for coin-agent flow

- [x] 2.1 Переписать `coin-executor/Jenkinsfile`: stages `Resolve version -> Test -> Publish agent image`
- [x] 2.2 Перевести сборку на `docker build -f Dockerfile.agent` (multi-stage, бинарь внутри образа)
- [x] 2.3 Передать в publish stage корректные credentials/env; убрать host `go build` + Maven publish

## 3. Publish-agent integration hardening

- [x] 3.1 Доработать `scripts/publish-agent.sh` при необходимости для стабильной работы из Jenkins workspace/job env
- [x] 3.2 Проверить, что после push создается draft `agent/coin-agent@<version>` через coin-api и promote не вызывается автоматически
- [x] 3.3 Улучшить диагностику ошибок publish/register (HTTP code + body) для быстрого ручного восстановления

## 4. Verification and docs

- [x] 4.1 Синхронизировать `coin-executor` в Gitea и обновить Jenkins job конфиг без автозапуска
- [ ] 4.2 Ручной прогон job: подтвердить наличие образа в Nexus Docker и draft версии в Platform UI/API
- [x] 4.3 Обновить runbook/README по новому потоку (job собирает image, promote выполняется вручную)
