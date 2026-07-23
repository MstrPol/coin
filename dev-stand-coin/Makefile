.PHONY: coin-lib push-lib upsert-jsl publish-executor publish-jnlp create-agent-test-job \
	push-demo-go-app create-demo-go-app-job demo-go-app

coin-lib: push-lib upsert-jsl

push-lib:
	./coin-lib/push-gitea.sh

upsert-jsl:
	./coin-lib/jenkins-global-lib.sh

# Builder: buildkit + coin-executor → Nexus
publish-executor:
	./coin-executor/publish.sh

# JNLP agent: mirror jenkins/inbound-agent → Nexus
publish-jnlp:
	./jnlp/mirror.sh

# Jenkins job coin-smoke (lib + dynamic agent)
create-agent-test-job:
	./smoke/create.sh

# Sample product repo → Gitea coin/demo-go-app
push-demo-go-app:
	./demo-go-app/push-gitea.sh

# Jenkins multibranch → Gitea coin/demo-go-app
create-demo-go-app-job:
	./demo-go-app/create-multibranch.sh

# Push sample + создать/пересканировать multibranch
demo-go-app: push-demo-go-app create-demo-go-app-job
