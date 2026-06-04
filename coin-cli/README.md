# coin-cli

CLI-утилита платформы Coin. Запускается в **CI agent** и выполняет логику pipeline: validate, version, run stages, render runtime-only Dockerfile.

Документация: [docs/agent-build-model.md](../docs/agent-build-model.md), [docs/architecture.md](../docs/architecture.md).

> **Новичок в Go?** Раздел «[Первый запуск](#первый-запуск)» написан специально для тебя.

---

## Содержание

- [Требования](#требования)
- [Первый запуск](#первый-запуск)
- [Сборка бинаря](#сборка-бинаря)
- [Запуск без сборки](#запуск-без-сборки)
- [Тесты](#тесты)
- [Команды CLI](#команды-cli)
- [Структура проекта](#структура-проекта)

---

## Требования

| Инструмент | Версия  | Установка |
|------------|---------|-----------|
| Go         | ≥ 1.22  | https://go.dev/dl/ |
| Git        | любая   | уже есть |

Проверить что Go установлен:

```bash
go version
# go version go1.22.x darwin/arm64
```

---

## Первый запуск

### 1. Клонировать монорепо Coin (если ещё не сделано)

```bash
git clone <url-монорепо>
cd coin/coin-cli
```

### 2. Скачать зависимости

Go хранит зависимости в модульном кеше. Один раз нужно их скачать:

```bash
go mod download
```

Это аналог `npm install` или `pip install -r requirements.txt`.  
После выполнения в директории появится `go.sum` — файл контрольных сумм (не удалять, коммитить в git).

### 3. Убедиться, что всё компилируется

```bash
go build ./...
```

Если ошибок нет — всё готово.

---

## Сборка бинаря

Platform CI: Jenkins job `coin-cli` → Nexus Maven (`maven-releases` / `maven-snapshots`, zip + classifier). Локально:

### Локально (для текущей ОС и архитектуры)

```bash
go build -o coin .
```

После этого в директории появится исполняемый файл `coin`:

```bash
./coin --help
```

### С указанием версии

При сборке CI передаёт версию через `-ldflags`:

```bash
go build -ldflags "-X coin.local/coin-cli/cmd.Version=1.2.3" -o coin .
```

### Кросс-компиляция (Linux amd64, например для CI-образа)

```bash
GOOS=linux GOARCH=amd64 go build -o coin_linux_amd64 .
```

Go компилирует бинарь для любой целевой платформы прямо с твоей машины — это встроенная возможность языка.

---

## Запуск без сборки

В Go можно запустить программу напрямую, без создания файла бинаря.  
Удобно при разработке — изменил код, сразу запустил:

```bash
# формат: go run . <аргументы>
go run . --help
go run . version
go run . validate --config path/to/.coin/config.yaml
go run . release bump --type patch --dry-run
```

`go run .` каждый раз компилирует код в памяти и сразу исполняет. Файл `coin` на диске **не создаётся**.

---

## Тесты

### Запустить все тесты

```bash
go test ./...
```

`./...` означает «текущий пакет и все вложенные рекурсивно».

### С подробным выводом (видно название каждого теста)

```bash
go test -v ./...
```

### Тесты конкретного пакета

```bash
go test ./internal/versioning/
go test ./internal/config/
go test ./cmd/
```

### Один конкретный тест

```bash
go test -v -run TestBump_Patch ./cmd/
go test -v -run TestCompute_ReleaseTag ./internal/versioning/
```

`-run` принимает регулярное выражение — можно указать часть имени:

```bash
go test -v -run TestCompute ./internal/versioning/
# запустит все TestCompute_*
```

### Покрытие кода

```bash
# текстовый отчёт в терминале
go test -cover ./...

# HTML-отчёт (откроется в браузере)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Пересборка и повторный запуск (без кеша)

Go кеширует результаты тестов. Если тест прошёл и код не менялся — он не перезапускается.  
Чтобы принудительно перезапустить:

```bash
go clean -testcache && go test ./...
```

---

## Команды CLI

После `go build -o coin .` (или через `go run .`):

```
coin validate                                   Проверить .coin/config.yaml
coin version                                    Показать текущую версию
coin version bump patch|minor|major             Создать следующий snapshot-тег
coin version bump patch|minor|major --type rc   Создать следующий RC-тег (только release/*)
coin version bump patch --dry-run               Показать тег без создания
coin run test                                   # native test (GP scripts/test.sh)
coin run build                                  # native compile + pack
coin run publish                                # publish
coin dockerfile render                          # runtime-only → .coin/generated/Dockerfile
```

Стадия `coin run build` для `*-app`: native compile в agent → render Dockerfile → `pack-image.sh`. См. [docs/agent-build-model.md](../docs/agent-build-model.md).

### `coin version`

Выводит последнюю версию из git-тегов. Нет тегов — `0.0.1`.

```bash
$ coin version
1.5.0-PROJ-404-rc-2          # HEAD помечен RC-тегом
0.0.1-PROJ-101-snapshot-2    # последний snapshot в репо
0.0.1                        # тегов нет (новый проект)
```

Используется в CI: `COIN_VERSION=$(coin version)`

### `coin version bump`

Создаёт следующий тег и пушит. По умолчанию `--type snapshot`.

```bash
coin version bump patch                  # → v0.0.1-PROJ-101-snapshot-1 (новая серия)
coin version bump patch                  # → v0.0.1-PROJ-101-snapshot-2 (продолжение)
coin version bump minor --type rc        # → v0.1.0-PROJ-404-rc-1 (только release/*)
coin version bump minor --type rc        # → v0.1.0-PROJ-404-rc-2 (итерация ПСИ)
coin version bump patch --dry-run        # показать тег без создания
```

Логика выбора базовой версии:
- Серия для текущего (JIRA-ID + тип) уже есть → продолжить (same base, N+1).
- Серии нет → взять последний base из репо + bump → N=1.

Флаги:

```
--type <snapshot|rc>   Тип тега (по умолчанию: snapshot)
--dry-run              Показать тег без создания
--config <path>        Путь к config.yaml (по умолчанию: .coin/config.yaml)
```

---

## Структура проекта

```
coin-cli/
├── main.go                        Точка входа — вызывает cmd.Execute()
├── go.mod                         Модуль и зависимости (аналог package.json)
├── go.sum                         Контрольные суммы зависимостей (не редактировать руками)
│
├── cmd/                           Cobra-команды (то, что вызывает пользователь)
│   ├── root.go                    Корневая команда `coin`
│   ├── validate.go                `coin validate`
│   ├── version.go                 `coin version` — вычисление версии и RC-тегирование
│   ├── version_test.go            Тесты вспомогательных функций
│   ├── run.go                     `coin run <stage>`
│   └── dockerfile.go              `coin dockerfile render`
│
├── internal/                      Внутренняя логика (недоступна снаружи модуля)
│   ├── config/
│   │   ├── config.go              Загрузка и валидация .coin/config.yaml
│   │   └── config_test.go
│   ├── versioning/
│   │   ├── versioning.go          Вычисление COIN_VERSION из git
│   │   └── versioning_test.go
│   ├── goldenpaths/               Golden paths из coin-golden-paths/
│   ├── starters/                  coin init — скелетоны из coin-starters/
│   ├── executor/
│   │   └── executor.go            Запуск стадий (preCommands, script, postCommands)
│   └── dockerfile/
│       └── render.go              Рендеринг Dockerfile из golden path
```

### Почему `internal/`?

В Go директория `internal` — специальная: пакеты внутри неё нельзя импортировать из других модулей.  
Это защита: логика `versioning`, `config` и т.д. не может быть случайно использована внешним кодом.

### Как добавить новую команду

1. Создать файл `cmd/mycommand.go` с `var myCmd = &cobra.Command{...}`
2. Зарегистрировать в `cmd/root.go`: `rootCmd.AddCommand(myCmd)`
3. Написать тесты в `cmd/mycommand_test.go`

Паттерн одинаковый для всех команд — посмотри `cmd/validate.go` как простейший пример.
