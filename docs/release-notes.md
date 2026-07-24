# Release Notes в Coin

Coin автоматически формирует release notes из git-истории и отправляет их во внутренний сервис QGM (Quality Gates Manager) через REST API.

---

## Концепция

Release notes — это структурированный документ, описывающий **что** вошло в конкретный дистрибутив:

- Jira-тикеты, связанные с релизом (smart-коммиты).
- Диапазон коммитов (первый и последний SHA).
- Участники разработки (по каждому тикету).
- Метаданные сборки (Jenkins job, ветка, время).

**Ключевое правило**: релиз в продакшн ≡ арtefact-тег в git ≡ release note в QGM.  
Пересборка невозможна без нового тега, а значит — без новой записи release notes.

---

## Как работают smart-коммиты

Coin сканирует сообщения git-коммитов в заданном диапазоне и извлекает Jira-тикеты по шаблону `[A-Z][A-Z0-9]+-\d+`.

```
feat(auth): добавить OAuth2 PROJ-123
fix: исправить NPE в обработчике MYTEAM-456 MYTEAM-457
```

Из таких коммитов в release notes попадут тикеты `PROJ-123`, `MYTEAM-456`, `MYTEAM-457`.  
Первая строка сообщения становится summary тикета.

---

## Конфигурация

Координаты артефакта задаются в `project:` — они используются во всех интеграциях:

```yaml
project:
  name: my-service              # → service name
  groupId: com.example.team     # → groupId в QGM
  artifactId: my-service        # → artifactId в QGM
```

| Поле | Обязательно | Описание |
|------|-------------|----------|
| `project.name` | **Да** | Имя сервиса. |
| `project.groupId` | **Да** | Домен команды (`groupId` в QGM). |
| `project.artifactId` | **Да** | Artifact ID (`artifactId` в QGM). |

**QGM в pipeline:** generate выполняется на каждый build (built-in stage ReleaseNotes).  
`coin-executor rn publish` вызывается только из `action: publish` и **отправляет JSON**, если задан `COIN_QGM_URL` (иначе skip — local pilot без QGM / без заглушки).

Platform env (не в product `.coin/config.yaml`):

| Env | Описание |
|-----|----------|
| `COIN_QGM_URL` | Base URL QGM API (например `http://qgm-stub:8080/qgm`). Пусто → skip publish. |
| `COIN_QGM_TOKEN` | Optional Bearer token. |
| `COIN_QGM_REPOSITORY` | Поле `repository` в payload (default `local`). |

Опционально для локального generate:

```yaml
# не поддерживается в product config v2 — только env выше
```

*Версия берётся из `COIN_VERSION` / `.coin/version.json` (next).*

---

## Команда `coin-executor rn generate`

Генерирует payload и сохраняет JSON в `.coin/temp/release-notes.json`.

```
coin-executor rn generate [флаги]
```

> Исторически команда называлась `coin rn generate` (coin-cli). После удаления CLI логика живёт в `coin-executor`.

### Авто-определение диапазона

Явно указывать `--from` **не нужно**. Команда сама определяет диапазон из модели ветвления:

1. Читает текущую версию (`coin version`), например `1.5.0-PROJ-404-rc-5`.
2. Извлекает `base = 1.5.0`.
3. Ищет последний RC-тег с **другим** base — это последний выпущенный релиз (`v1.4.2-PROJ-300-rc-3`).
4. Все коммиты от `v1.4.2-PROJ-300-rc-3` до HEAD включаются в release notes.

Результат: если текущий релиз на `rc-5`, в RN попадают **все** Jira-тикеты начиная с `rc-1` этой серии. Флаг RC-номера не влияет на состав release notes.

| Ситуация | `--from` авто |
|----------|---------------|
| Первый релиз проекта | нет нижней границы — вся история |
| Второй и последующие | последний RC-тег предыдущего base |

### Флаги

| Флаг             | По умолчанию                    | Описание |
|------------------|---------------------------------|----------|
| `--release-link` | (пусто)                         | Ссылка на Jira-задачу «Release 2.0». |
| `--output`       | `.coin/temp/release-notes.json` | Путь для сохранения JSON. |
| `--config`       | `.coin/config.yaml`             | Путь до конфига. |
| `--dry-run`      | false                           | Показать сводку без сохранения на диск. |

### Примеры

```bash
# Стандартный запуск — всё определяется автоматически
coin-executor rn generate

# Убедиться что попало до записи файла
coin-executor rn generate --dry-run

# С указанием задачи на релиз (Jira «Release 2.0»)
coin-executor rn generate --release-link https://jira.example.com/browse/PROJ-404
```

---

## Структура JSON

Файл `.coin/temp/release-notes.json` соответствует схеме `ReleaseNoteApiRequest` QGM API.

```json
{
  "repository": "Nexus_PROD",
  "groupId": "com.example.team",
  "artifactId": "my-service",
  "version": "1.5.0-PROJ-404-rc-3",
  "releaseNotes": [
    { "issue": "PROJ-401", "summary": "feat: добавить экспорт отчётов" },
    { "issue": "PROJ-402", "summary": "fix: исправить пагинацию" }
  ],
  "releaseLink": "https://jira.example.com/browse/PROJ-404",
  "codeNotes": [
    {
      "commit": "f401bba58c95ffbd510bdb590c9f6d2d538f497d",
      "repository": "ssh://git@bitbucket.example.com/team/my-service.git",
      "from":   "33d01f71197d8d2dc588b49de9b714953e3bae28"
    }
  ],
  "buildInfo": {
    "meta": [],
    "buildNumber": "42",
    "buildUrl": "https://jenkins.example.com/job/my-service/42/",
    "branchName": "release/PROJ-404",
    "jobName": "my-service/release/PROJ-404"
  },
  "meta": [
    { "key": "coin.version", "value": "1.5.0-PROJ-404-rc-3", "display": "Coin Version" },
    { "key": "generated.at", "value": "2026-06-02T09:00:00Z", "display": "Generated At" }
  ],
  "links": [],
  "contributors": {
    "PROJ-401": [{ "userName": "Иван Иванов", "email": "ivanov@example.com" }],
    "PROJ-402": [{ "userName": "Пётр Петров", "email": "petrov@example.com" }]
  },
  "content": {}
}
```

---

## Место в пайплайне

Built-in stage **ReleaseNotes** (после Version) запускает `coin-executor rn generate` на **каждый** build и архивирует `.coin/temp/release-notes.json`.

Отправка в QGM — только при `action: publish` (`coin-executor rn publish` / вызов из executor). Без `COIN_QGM_URL` — skip.

| Шаг | Команда | Когда |
|-----|---------|-------|
| 1. Plan version | `coin-executor plan-version` | каждый build (Version) |
| 2. Сгенерировать RN | `coin-executor rn generate` | каждый build (ReleaseNotes) |
| 3. Проверить тикеты | артефакт `.coin/temp/release-notes.json` | тимлид / релиз-менеджер |
| 4. Отправить в QGM | `coin-executor rn publish` | только publish + `COIN_QGM_URL` |

> Заглушка QGM для local pilot опциональна: задайте `COIN_QGM_URL` на stub, иначе шаг 4 только логирует skip.

---

## Временное хранилище `.coin/temp/`

Папка `.coin/temp/` — это **локальный буфер** для артефактов, созданных в процессе сборки.  
Она не коммитится в репозиторий (добавлена в `.gitignore`).

Содержимое папки:

| Файл | Когда создаётся | Описание |
|------|-----------------|----------|
| `release-notes.json` | `coin-executor rn generate` | Payload для QGM API |

---

## FAQ

**Почему release notes не из Jira, а из git?**

Git — источник истины для того, что реально вошло в дистрибутив. Jira может содержать тикеты, для которых нет ни одного коммита. Smart-коммиты гарантируют, что в RN попадут только тикеты с реальными изменениями кода.

**Что если коммит не содержит Jira-тикет?**

Коммит будет учтён в `codeNotes` (диапазон SHA), но в `releaseNotes` не попадёт. Рекомендуем придерживаться правила: каждый коммит в `release/*` должен содержать хотя бы один Jira-тикет.

**Что меняется при переходе от rc-1 к rc-5?**

Ничего — диапазон определяется автоматически от предыдущего **base**, а не от предыдущего snapshot/RC. Для snapshot-1 … snapshot-N одного и того же base `coin-executor rn generate` вернёт одинаковый список тикетов (плюс новые). Запускать команду заново при каждом build — нормально: файл обновится.

**Как проверить результат перед отправкой?**

```bash
coin-executor rn generate --dry-run
```

Или посмотреть файл напрямую:

```bash
cat .coin/temp/release-notes.json | python3 -m json.tool
```
