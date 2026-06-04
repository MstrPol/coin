FROM jenkins/jenkins:lts-jdk17

USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends git curl docker.io docker-cli ca-certificates zip \
    && rm -rf /var/lib/apt/lists/* \
    && usermod -aG docker jenkins

COPY plugins.txt /usr/share/jenkins/ref/plugins.txt
# updates.jenkins.io/CDN иногда обрывает update-center.json (Truncated chunk) — retry
RUN set -eux; \
    for attempt in $(seq 1 5); do \
      echo "==> jenkins-plugin-cli attempt ${attempt}/5"; \
      if jenkins-plugin-cli --verbose --plugin-file /usr/share/jenkins/ref/plugins.txt; then \
        exit 0; \
      fi; \
      echo "==> plugin download failed, retry in 20s..."; \
      sleep 20; \
    done; \
    echo "jenkins-plugin-cli failed after 5 attempts" >&2; \
    exit 1

COPY casc.yaml /usr/share/jenkins/ref/casc.yaml
COPY jenkins-entrypoint.sh /usr/local/bin/jenkins-entrypoint.sh
RUN chmod +x /usr/local/bin/jenkins-entrypoint.sh

ENV JAVA_OPTS="-Djenkins.install.runSetupWizard=false"

USER root
ENTRYPOINT ["/usr/local/bin/jenkins-entrypoint.sh"]
