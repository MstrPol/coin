# Чеклист миграции repo (Wave 1)

> **⚠️ Corp gate:** применять только после доступа в corp. Сейчас — проверка на **`samples/demo-go-app`**.

**Ticket:** P1-07  
**Scope:** Wave 1 — до **50 repos** на `go-app@1.0.0`.  
**Rollback на config v1 запрещён.**

Полный runbook rollout: [wave-1-migration.md](../runbooks/wave-1-migration.md).

## Перед началом (repo owner)

- [ ] Сервис на **go-app** (v1 template или greenfield)
- [ ] Есть Jenkins multibranch / pipeline job
- [ ] Credential `nexus-docker` (или ID из manifest) существует
- [ ] Platform подтвердила: coin-api + Nexus cache + executor опубликованы

## Шаг 1 — Inventory

| Поле | Значение |
|------|----------|
| Repo URL | |
| Текущий GP | |
| Целевой | `go-app` / `1.0.0` |
| Jenkins job | |
| Owner | |

## Шаг 2 — Config v2

- [ ] Файл `.coin/config.yaml` в корне repo
- [ ] Секция `coin.goldenPath` + `coin.version` (не `template`/`templateVersion`)
- [ ] `project.name`, `project.groupId`, `jenkins.credentials.docker`

Пример и mapping v1→v2: [migrate-config-v1-to-v2.md](migrate-config-v1-to-v2.md).

## Шаг 3 — Jenkinsfile

- [ ] Удалён v1 Shared Library pipeline (`coinPipeline()`)
- [ ] Содержимое = [`Jenkinsfile.coin`](../../coin-starters/Jenkinsfile.coin)
- [ ] (Optional) env `COIN_API_URL`, `COIN_MANIFEST_CACHE_BASE`

## Шаг 4 — PR / merge

- [ ] PR прошёл review platform (или checklist self-service для Wave 1)
- [ ] Merge в default branch (`main`/`master`)

## Шаг 5 — Jenkins E2E

- [ ] Scan multibranch → новый commit обнаружен
- [ ] **Resolve manifest** — green (API или Nexus fallback)
- [ ] **Validate → Test → Build** — green
- [ ] **Report** — green, `✓ build report sent`
- [ ] Docker image в registry

## Шаг 6 — Verify report (platform / optional owner)

```sql
SELECT p.name, br.result, br.build_url, br.reported_at
FROM build_reports br
JOIN projects p ON p.id = br.project_id
WHERE p.name = '<service-name>'
ORDER BY br.id DESC LIMIT 1;
```

- [ ] Row с `result=success` и корректным `build_url`

## Шаг 7 — Sign-off

| | |
|---|---|
| Дата | |
| Build URL | |
| Мигрировал | |
| Batch (canary / 1 / 2) | |
| Заметки | |

## После миграции

- [ ] Удалить legacy Jenkins shared library refs из README repo (если были)
- [ ] Обновить tracker: repo → **v2** ✅
- [ ] (Wave 3+) Настроить canary mode при rollout — см. [canary.md](../canary.md)

## Не мигрировать в Wave 1

| Условие | Действие |
|---------|----------|
| GP ≠ go-app (java, python, custom) | Wave 2 |
| Custom Groovy hooks в Jenkinsfile | Platform review — вынести в executor или отложить |
| Нет K8s cloud в Jenkins | Сначала [jenkins-setup.md](../jenkins-setup.md) |

## Troubleshooting

См. [troubleshoot-ci.md](troubleshoot-ci.md).

## Waves дальше

| Wave | Repos | GP | Документ |
|------|-------|-----|----------|
| 1 | 50 | go-app@1.0.0 | этот чеклист |
| 2 | 500 | + java/python | TBD Wave 2 runbook |
| 3 | 1500+ | full fleet | [wave-3-migration.md](../runbooks/wave-3-migration.md) |
