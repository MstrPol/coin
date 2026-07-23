.PHONY: up gitea-up nexus-up k3s-up jenkins-up dashboard-up

up: nexus-up gitea-up k3s-up jenkins-up

gitea-up:
	docker compose up -d gitea
	./gitea/prepare.sh

nexus-up:
	docker compose up -d nexus
	./nexus/prepare.sh

k3s-up:
	chmod +x ./k3s/sync-registries.sh ./k3s/prepare.sh
	./k3s/sync-registries.sh
	docker compose up -d k3s
	./k3s/prepare.sh

dashboard-up:
	chmod +x ./k3s/dashboard-up.sh
	./k3s/dashboard-up.sh

jenkins-up:
	chmod +x ./jenkins/setup-k8s-auth.sh ./jenkins/prepare.sh
	./jenkins/setup-k8s-auth.sh
	docker compose up -d --build jenkins
	./jenkins/prepare.sh
