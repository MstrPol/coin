# ADR: Build Engine Contract

## Статус

accepted

> **Operational SoT:** текущая CI runtime-модель (agent, bootstrap, engines, publish) — [coin-ci-runtime](coin-ci-runtime.md).  
> Этот ADR фиксирует **решение** о введении `build.engine` и hard cut; секция «Контекст» описывает **pre-hard-cut** состояние.

## Контекст (pre-hard-cut, 2026-06)

До hard cut Jenkins runtime model использовала динамический pod с stack image, где установлен языковой toolchain. Для `go-app` это приводило к `go build` внутри agent и последующему `docker build`.

Проблемы:

- agent images становятся language-specific;
- кеши language build и Docker layers не являются стабильной частью платформенного контракта;
- `/var/run/docker.sock` в pod является security risk;
- `coin-lib` остается тонкой, но build behavior фактически зависит от содержимого agent image.

## Решение

Вводим build engine contract:

```yaml
build:
  engine: buildkit | dockerfile
```

Источник SoT: `coin-gp-content/stacks/<gp>/content.yaml` (schema v2).

> **2026-06 update:** buildpack hard cut; BYO `dockerfile` — отдельный GP (`go-app-docker`), не managed Containerfile duplicate.

Resolved manifest обязан содержать выбранный `build` object. `coin-executor` выбирает implementation по `manifest.build.engine`.

`coin-lib` не интерпретирует build engine и не получает build business logic.

## Hard Cut

Не мигрируем старую модель и не поддерживаем dual path.

Сразу перестраиваем runtime:

- universal `coin-agent` без Go/Java/Python/Node toolchains;
- `coin-agent` собирается на базе `jenkins/inbound-agent` и содержит `coin-executor`, `podman`, `buildctl`, `buildkitd` и registry tools;
- pod содержит один container `jnlp`; отдельный `stack` container и `manifest.jnlp` удаляются;
- `coin-agent` собирается и публикуется из `coin-executor/`, потому что это runtime-упаковка executor-а;
- `coin-jenkins-agents/` полностью удаляется как superseded компонент;
- BuildKit как основной managed build path; BYO Dockerfile — отдельный GP profile;
- для local pilot `buildkitd` **не** стартует в bootstrap на arm64; `podman system service` — обязательный bootstrap step;
- default engine для `go-app` — `buildkit`;
- Docker socket хоста не используется; container builds через `podman system service` внутри agent pod;
- произвольные GP `scripts/*.sh` не являются runtime build path;
- project config не задает `build.engine` в первой итерации.

Engine `dockerfile` остается как explicit engine внутри нового контракта. Он не является legacy path вне `build.engine`.

Stage behavior переносится из GP shell scripts в typed executor/build-engine contract:

| Старый подход | Новый подход |
|---------------|--------------|
| `scripts/validate.sh` | built-in `coin-executor validate` + optional BuildKit target `validate` |
| `scripts/test.sh` | BuildKit target `test` или typed engine policy |
| `scripts/build.sh` | `coin-executor build` dispatch по `build.engine` |
| `scripts/publish.sh` | built-in `coin-executor publish` по `.coin/outputs.json` |

Новый произвольный shell hook требует отдельного ADR.

## Последствия

Плюсы:

- agent image становится универсальным;
- language toolchain переносится в builder images/Dockerfile stages;
- build cache можно хранить в registry;
- non-container artifacts можно получать через BuildKit local output;
- `coin-lib` сохраняет роль Jenkins glue.
- ownership agent image совпадает с ownership executor runtime.

Минусы:

- нужно расширить manifest contract;
- `coin-executor` получает build dispatch;
- GP content должен описывать build engine и target/output policy;
- local pilot должен запускать `podman system service` в bootstrap; для engine `buildkit` на **arm64 pilot** container builds идут через **podman build** (buildctl RUN несовместим с nested runc в k3s);
- на **amd64 corp** (roadmap) — нативный `buildkitd` + `buildctl` без podman-fallback;
- существующие GP `scripts/*.sh` удаляются или перестают быть частью runtime manifest.
- `coin-jenkins-agents/`, его Jenkins jobs, scripts, docs и seed references должны быть удалены из target tree.

## Отклонённые Альтернативы

### Language-specific agents

Отклонено: ведет к большому набору images, холодным кешам и размытому build contract.

### Arbitrary GP shell scripts

Отклонено: возвращает скрытую зависимость от содержимого agent image и размывает typed build contract.

### Project-level engine selection

Отклонено для первой итерации: build engine является политикой Golden Path. Project override требует отдельного ADR.

### Buildpacks only

Отклонено: хорошо для OCI images, но хуже для non-container artifacts и сложных multi-output build flows.

### BuildKit only

Отклонено: Buildpacks дают более простой default для типовых app images.

### Dual-container pod (jnlp + stack)

Отклонено: два образа усложняют composition, pod template и `container('stack')` switching без выигрыша после перехода на custom inbound-agent image.
