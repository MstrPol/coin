# coin-cli

CLI-утилита платформы Coin. Запускается внутри CI-агента и выполняет всю полезную логику сборки: валидацию конфига, вычисление версии, запуск стадий, рендеринг Dockerfile и управление тегами релизов.

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
coin validate                           Проверить .coin/config.yaml
coin version                            Вычислить COIN_VERSION из git
coin run test                           Запустить стадию тестирования
coin run build                          Запустить стадию сборки
coin run publish                        Запустить стадию публикации
coin dockerfile render                  Сгенерировать .coin/generated/Dockerfile
coin release bump --type patch          Создать тег следующего patch-релиза
coin release bump --type minor          Создать тег следующего minor-релиза
coin release bump --type major          Создать тег следующего major-релиза
coin release bump --type rc             Создать следующий release candidate тег
coin release bump --type patch --dry-run  Показать тег без создания
```

Флаги, общие для большинства команд:

```
--config <path>   Путь к config.yaml (по умолчанию: .coin/config.yaml)
--dry-run         Показать что будет сделано, не делая этого (только для release bump)
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
│   ├── version.go                 `coin version`
│   ├── run.go                     `coin run <stage>`
│   ├── dockerfile.go              `coin dockerfile render`
│   ├── release.go                 `coin release bump`
│   └── release_test.go            Тесты чистых функций release-логики
│
├── internal/                      Внутренняя логика (недоступна снаружи модуля)
│   ├── config/
│   │   ├── config.go              Загрузка и валидация .coin/config.yaml
│   │   └── config_test.go
│   ├── versioning/
│   │   ├── versioning.go          Вычисление COIN_VERSION из git
│   │   └── versioning_test.go
│   ├── executor/
│   │   └── executor.go            Запуск стадий (preCommands, script, postCommands)
│   └── dockerfile/
│       └── render.go              Рендеринг Dockerfile из embedded-шаблонов
│
└── embed/
    ├── embed.go                   go:embed — встраивает scripts/ и dockerfiles/ в бинарь
    ├── scripts/                   Shell-скрипты для стадий (test/build/publish по стекам)
    └── dockerfiles/               Шаблоны Dockerfile по стекам
```

### Почему `internal/`?

В Go директория `internal` — специальная: пакеты внутри неё нельзя импортировать из других модулей.  
Это защита: логика `versioning`, `config` и т.д. не может быть случайно использована внешним кодом.

### Как добавить новую команду

1. Создать файл `cmd/mycommand.go` с `var myCmd = &cobra.Command{...}`
2. Зарегистрировать в `cmd/root.go`: `rootCmd.AddCommand(myCmd)`
3. Написать тесты в `cmd/mycommand_test.go`

Паттерн одинаковый для всех команд — посмотри `cmd/validate.go` как простейший пример.
