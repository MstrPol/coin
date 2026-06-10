.PHONY: up bootstrap down reset build-jenkins dashboard coin-platform coin-jenkins-agents coin-starters coin-executor agents-build samples e2e-mvp1 postgres-up coin-api-up coin-ui-up endpoints scan-fleet scan-cronjob-apply scan-cronjob-run

up:
	docker compose up -d

postgres-up:
	docker compose up -d postgres

coin-api-up:
	docker compose up -d postgres coin-api

coin-ui-up:
	docker compose up -d coin-api coin-ui

openapi-ui:
	cd ../coin-ui && $(MAKE) openapi-ui

scan-fleet:
	cd ../coin-api && \
	  DATABASE_URL=postgres://$${POSTGRES_USER:-coin}:$${POSTGRES_PASSWORD:-coin}@localhost:$${POSTGRES_PORT:-5432}/$${POSTGRES_DB:-coin}?sslmode=disable \
	  GITEA_URL=http://localhost:$${GITEA_HTTP_PORT:-3000} \
	  GITEA_USER=$${GITEA_USER:-coin} \
	  GITEA_PASSWORD=$${GITEA_PASSWORD:-coin} \
	  GITEA_ORG=$${GITEA_ORG:-coin} \
	  go run ./cmd/scanner

scan-cronjob-apply:
	docker compose exec -T k3s kubectl apply -f /k3s/coin-scanner-cronjob.yaml

scan-cronjob-run:
	docker compose exec -T k3s kubectl delete job coin-fleet-scanner-manual --ignore-not-found
	docker compose exec -T k3s kubectl apply -f /k3s/coin-scanner-job.yaml
	docker compose exec -T k3s kubectl wait --for=condition=complete job/coin-fleet-scanner-manual --timeout=7200s
	docker compose exec -T k3s kubectl logs job/coin-fleet-scanner-manual

endpoints:
	chmod +x scripts/register-stack-endpoints.sh scripts/fix-k3s.sh
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
	chmod +x scripts/coin-platform.sh scripts/coin-jenkins-agents.sh scripts/coin-starters.sh scripts/lib/common.sh
	./scripts/coin-platform.sh

coin-jenkins-agents:
	chmod +x scripts/coin-jenkins-agents.sh scripts/lib/common.sh
	./scripts/coin-jenkins-agents.sh

coin-starters:
	chmod +x scripts/coin-starters.sh scripts/lib/common.sh
	./scripts/coin-starters.sh

coin-executor:
	chmod +x scripts/coin-executor.sh scripts/lib/common.sh
	./scripts/coin-executor.sh

agents-build:
	chmod +x scripts/agents-build.sh scripts/lib/common.sh
	./scripts/agents-build.sh

samples:
	chmod +x scripts/samples.sh scripts/lib/common.sh
	./scripts/samples.sh

e2e-mvp1:
	chmod +x scripts/e2e-mvp1.sh scripts/lib/common.sh
	./scripts/e2e-mvp1.sh
