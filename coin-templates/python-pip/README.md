# Python + pip (Coin template)

Минимальный шаблон сервиса на Python с `requirements.txt`.

- Dockerfile и `.dockerignore` генерирует Coin.
- Стандартные test/build/publish сценарии живут в `coin-lib`.
- Расширения задаются в `.coin/config.yaml` через `preCommands`, `commands`, `postCommands`.
