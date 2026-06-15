# coin-gp-content

Immutable GP content packages: scripts, Dockerfile, validate schema per golden path stack.

## Layout

```
stacks/go-app/
  content.yaml
  scripts/
  dockerfiles/
  schemas/
```

Publish flow: next-version → ZIP → Nexus → coin-api register → artifact bodies.

## Local publish

```bash
./scripts/publish-content.sh go-app 1.0.0
```

Zip → Nexus `maven-releases` (`coin/gp-content/{name}/{ver}/`) → register `gp-content/go-app@1.0.0` в coin-api.

## CI

Jenkins job `coin-gp-content` (см. `docker/scripts/coin-gp-content.sh`).
