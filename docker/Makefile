.PHONY: up bootstrap down reset build-jenkins dashboard coin-lib coin-platform coin-cli agents-build samples

up:
	docker compose up -d

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

coin-lib:
	chmod +x scripts/coin-lib.sh scripts/lib/common.sh
	./scripts/coin-lib.sh

coin-platform:
	chmod +x scripts/coin-platform.sh scripts/lib/common.sh
	./scripts/coin-platform.sh

coin-cli:
	chmod +x scripts/coin-cli.sh scripts/lib/common.sh
	./scripts/coin-cli.sh

agents-build:
	chmod +x scripts/agents-build.sh scripts/lib/common.sh
	./scripts/agents-build.sh

samples:
	chmod +x scripts/samples.sh scripts/lib/common.sh
	./scripts/samples.sh
