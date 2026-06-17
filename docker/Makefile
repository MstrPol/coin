.PHONY: up bootstrap down reset build-jenkins dashboard coin-platform coin-starters coin-executor coin-gp-content coin-lib publish-agent wipe-gp wipe-e2e-fresh samples e2e-mvp1 e2e-bootstrap e2e-demo-go-app e2e-build-engines postgres-up coin-api-up coin-ui-up endpoints seed-jenkins-lib e2e-jenkins-lib

up:
	docker compose up -d

postgres-up:
	docker compose up -d postgres

build-coin-api:
	DOCKER_BUILDKIT=1 docker compose --progress plain build coin-api

coin-api-up: build-coin-api
	docker compose up -d --force-recreate postgres coin-api

build-coin-ui:
	DOCKER_BUILDKIT=1 docker compose --progress plain build coin-ui

coin-ui-up: build-coin-ui
	docker compose up -d --force-recreate coin-api coin-ui

openapi-ui:
	cd ../coin-ui && $(MAKE) openapi-ui

endpoints:
	chmod +x scripts/sync-k3s-registries.sh scripts/register-stack-endpoints.sh scripts/fix-k3s.sh
	./scripts/sync-k3s-registries.sh
	./scripts/register-stack-endpoints.sh

build-jenkins:
	DOCKER_BUILDKIT=1 docker compose --progress plain build jenkins

dashboard:
	chmod +x scripts/setup-dashboard.sh
	./scripts/setup-dashboard.sh

bootstrap:
	chmod +x scripts/*.sh scripts/lib/*.sh jenkins/jenkins-entrypoint.sh
	./scripts/bootstrap.sh

down:
	chmod +x scripts/teardown.sh
	./scripts/teardown.sh

reset:
	chmod +x scripts/reset.sh
	./scripts/reset.sh

coin-platform:
	chmod +x scripts/coin-platform.sh scripts/coin-starters.sh scripts/lib/common.sh
	./scripts/coin-platform.sh

coin-starters:
	chmod +x scripts/coin-starters.sh scripts/lib/common.sh
	./scripts/coin-starters.sh

coin-executor:
	chmod +x scripts/coin-executor.sh scripts/lib/common.sh
	./scripts/coin-executor.sh

coin-gp-content:
	chmod +x scripts/coin-gp-content.sh scripts/lib/common.sh
	./scripts/coin-gp-content.sh

coin-lib:
	chmod +x scripts/coin-lib.sh scripts/lib/common.sh
	./scripts/coin-lib.sh

coin-lib-http:
	chmod +x scripts/coin-lib-http.sh scripts/lib/common.sh
	./scripts/coin-lib-http.sh

publish-agent:
	chmod +x ../coin-executor/scripts/publish-agent.sh
	VERSION=$${VERSION:-1.0.0} GOARCH=$${GOARCH:-amd64} ../coin-executor/scripts/publish-agent.sh "$${VERSION}" "$${GOARCH}"

wipe-gp:
	chmod +x scripts/wipe-gp.sh
	./scripts/wipe-gp.sh

wipe-e2e-fresh:
	chmod +x scripts/wipe-e2e-fresh.sh scripts/lib/common.sh
	./scripts/wipe-e2e-fresh.sh

samples:
	chmod +x scripts/samples.sh scripts/lib/common.sh
	./scripts/samples.sh

e2e-mvp1:
	chmod +x scripts/e2e-mvp1.sh scripts/lib/common.sh
	./scripts/e2e-mvp1.sh

seed-jenkins-lib:
	chmod +x scripts/seed-jenkins-lib-stack.sh scripts/lib/common.sh
	./scripts/seed-jenkins-lib-stack.sh

e2e-bootstrap:
	chmod +x scripts/e2e-bootstrap.sh scripts/lib/common.sh
	./scripts/e2e-bootstrap.sh

e2e-jenkins-lib:
	chmod +x scripts/e2e-jenkins-lib.sh scripts/lib/common.sh
	./scripts/e2e-jenkins-lib.sh

e2e-demo-go-app:
	chmod +x scripts/e2e-demo-go-app.sh scripts/lib/common.sh
	./scripts/e2e-demo-go-app.sh

e2e-build-engines:
	chmod +x scripts/e2e-build-engines.sh scripts/e2e-demo-go-app.sh scripts/prune-k3s-disk.sh scripts/lib/common.sh
	./scripts/e2e-build-engines.sh
