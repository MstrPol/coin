# ADR: Pipeline bundle layer

> **Статус: superseded** (2026-06-10)  
> Заменено: pinned **coin-lib** Shared Library + **gp-content** — [jenkins-lib-http-nexus](jenkins-lib-http-nexus.md), [coin-ci-runtime](coin-ci-runtime.md).

**Статус (исторический):** accepted  
**Дата:** 2026-06-10

## Контекст

Jenkins Shared Library создаёт drift и не даёт immutable pin. Jenkins-specific логика (pod, creds, stages) нужна, но не в продуктовом repo.

## Решение

1. **pipeline-bundle** — versioned zip в Nexus, registry component `pipeline-bundle/{stack}@semver`
2. Содержимое: Groovy modules (kernel, pod, creds, params, stages), shell scripts, Dockerfile template, validate schema
3. Platform CI: zip → Nexus → coin-api register (+ artifact bodies)
4. Manifest: `jnlp`, `orchestration.bundle`, `pipeline`, `validateSchema`, `dockerfileTemplate`
5. Product `Jenkinsfile.coin`: resolve → download/verify/unzip bundle → `load entrypoint`
6. **coin-executor** остаётся execution engine; не владеет Jenkins DSL

## Q1–Q5

| # | Решение |
|---|---------|
| Q1 | Pin на GP release как `pipeline-bundle` component |
| Q2 | Orchestration в bundle; env fallback только migration |
| Q3 | Expand manifest v1 при resolve |
| Q4 | Platform CI `coin-pipeline-bundles` |
| Q5 | Superseded: Jenkins credentials не входят в manifest; источник — product config / coin-lib defaults |

## Отклонённые альтернативы

- Jenkins Shared Library (в т.ч. pinned by commit SHA в runtime)
- Один monolithic groovy без bundle structure
- Orchestration только в coin-starters git
