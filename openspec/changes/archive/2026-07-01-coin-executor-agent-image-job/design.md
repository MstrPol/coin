## Context

В local pilot пайплайн `coin-executor` сейчас публикует только бинарь `coin-executor` в Maven repository, а сборка/публикация `coin-agent` образа и регистрация `agent` draft выполняются отдельно через ручной запуск скрипта. Это увеличивает количество ручных шагов перед итоговым E2E и создает расхождение между ожиданием от CI job и фактическим runtime flow.

Текущие инварианты проекта сохраняются:
- runtime определяется pin `agent` в GP composition;
- promote `agent` остается ручным;
- build-логика живет в `coin-executor` скриптах, Jenkinsfile выполняет orchestration.

## Goals / Non-Goals

**Goals:**
- Перевести job `coin-executor` на целевой поток `coin-agent`: build image, push в Nexus Docker, register draft в coin-api.
- Сохранить Jenkinsfile тонким: параметры + credentials binding + вызов `publish-agent.sh`.
- Сохранить manual gate (`draft -> published`) без auto-promote.
- Обеспечить совместимость с arm64/amd64 через параметризацию `GOARCH`.

**Non-Goals:**
- Изменение контрактов `coin-api` для agent draft/promote.
- Изменение модели GP composition и runtime pinning.
- Введение новой Jenkins shared library бизнес-логики.
- Fleet rollout / corp-специфичные ветки выполнения.

## Decisions

1. **Job строит и публикует именно `coin-agent`, а не standalone binary**
   - **Почему:** продуктовый runtime теперь определяется `agent` image, а не отдельным `executor` component.
   - **Альтернатива:** оставить current binary-only job + отдельный ручной шаг `make publish-agent`.
   - **Почему не выбрана:** лишний ручной шаг и выше риск пропуска критичного этапа перед E2E.

2. **Использовать `scripts/publish-agent.sh` как единую точку бизнес-логики**
   - **Почему:** соблюдает границу Jenkins/runtime (Jenkins orchestration, доменная логика в repo scripts).
   - **Альтернатива:** реализовать docker build/push и API register inline в Jenkinsfile.
   - **Почему не выбрана:** усложняет Jenkinsfile, повышает дублирование и риск расхождения со скриптовым контуром.

3. **Сохранить manual promote gate после draft register**
   - **Почему:** соответствует действующему requirement `runtime-agent-registry` и UI-first flow.
   - **Альтернатива:** auto-promote в том же job.
   - **Почему не выбрана:** ломает контроль качества enabling team и текущие guardrails.

4. **Параметризовать версию через semver bump в job и передавать ее в publish-agent**
   - **Почему:** сохраняет текущий UX job (`BUMP`) и предсказуемый version flow.
   - **Альтернатива:** фиксированная версия как input string без bump.
   - **Почему не выбрана:** больше ручных ошибок, нет единой политики инкремента.

5. **Сборка образа через multi-stage `Dockerfile.agent`, а не host `go build` + COPY binary**
   - **Почему:** единый reproducible pipeline, меньший runtime-слой, бинарь и runtime собираются в одном `docker build`.
   - **Структура:**
     - `executor-builder` (`golang`) → `go build coin-executor`
     - `buildkit-bin` (`moby/buildkit`) → бинарники buildkit
     - runtime (`jenkins/inbound-agent`) → podman + buildkit + baked binary
   - **Альтернатива:** собирать бинарь на Jenkins node и копировать в slim context.
   - **Почему не выбрана:** дублирование логики сборки, расхождение dev/CI, лишний stash/unstash.

## Risks / Trade-offs

- **[Risk]** Jenkins runner может не иметь доступа к docker daemon для сборки образа.  
  **Mitigation:** явная проверка `docker version`/ошибка в ранней стадии и документация prereqs.

- **[Risk]** Переход на image-first job убирает публикацию бинаря, который может использоваться внешними ручными сценариями.  
  **Mitigation:** зафиксировать в runbook новый source of truth и при необходимости оставить отдельный legacy job.

- **[Risk]** Ошибки в Nexus credentials или API key приводят к частично успешному run (image pushed, draft не создан).  
  **Mitigation:** fail-fast после API register, выводить URL image/version в логах для ручного восстановления.

- **[Trade-off]** Job становится тяжелее и дольше (docker build + push), чем binary-only pipeline.  
  **Mitigation:** использовать кэш docker layers и держать отдельные стадии с понятной диагностикой.

## Migration Plan

1. Обновить `coin-executor/Jenkinsfile` под этапы image publish + draft register.
2. Проверить наличие требуемых credentials в Jenkins (`nexus-docker`, `coin-publisher-api-key`).
3. Синхронизировать репозиторий в Gitea и reload Jenkins job через текущий bootstrap flow.
4. Прогнать ручной build параметрами (`BUMP`, `GOARCH`) и убедиться, что:
   - образ появился в Nexus Docker;
   - draft `agent/coin-agent@<version>` создан в coin-api.
5. Оставить promote ручным в Platform UI.
6. При необходимости rollback: вернуть предыдущий Jenkinsfile из git history и пересинхронизировать job.
