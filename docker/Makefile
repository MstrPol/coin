.PHONY: up bootstrap down demo repos logs gitea dashboard k8s-auth k8s-endpoints k3s-fix dashboard-token platform-build

up:
	docker compose up -d

bootstrap:
	chmod +x scripts/*.sh jenkins/jenkins-entrypoint.sh
	./scripts/bootstrap.sh

down:
	chmod +x scripts/teardown.sh
	./scripts/teardown.sh

gitea:
	chmod +x scripts/prepare-gitea.sh
	./scripts/prepare-gitea.sh

dashboard:
	chmod +x scripts/setup-dashboard.sh
	./scripts/setup-dashboard.sh

k8s-auth:
	chmod +x scripts/setup-jenkins-k8s-auth.sh
	./scripts/setup-jenkins-k8s-auth.sh

k8s-endpoints:
	chmod +x scripts/register-jenkins-k8s-endpoints.sh
	./scripts/register-jenkins-k8s-endpoints.sh

k3s-fix:
	chmod +x scripts/fix-k3s.sh
	./scripts/fix-k3s.sh

dashboard-token:
	@token=$$(docker compose exec -T k3s kubectl -n kubernetes-dashboard create token admin-user --duration=24h | tr -d '\r\n'); \
	printf '%s\n' "$$token"; \
	if command -v pbcopy >/dev/null 2>&1; then printf '%s' "$$token" | pbcopy && echo "→ скопировано в буфер обмена (macOS)"; fi; \
	echo "→ вставьте в Dashboard: https://localhost:$${K3S_DASHBOARD_PORT:-8443} (Token, без пробелов)"

repos:
	chmod +x scripts/push-coin.sh scripts/prepare-demo.sh
	./scripts/push-coin.sh
	./scripts/prepare-demo.sh

platform-build:
	chmod +x scripts/trigger-platform-build.sh
	./scripts/trigger-platform-build.sh

demo: repos

logs:
	docker compose logs -f --tail=200
