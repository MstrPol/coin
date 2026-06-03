FROM jenkins/jenkins:lts-jdk17

USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends git curl docker.io ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && usermod -aG docker jenkins

# Go для platform job coin-cli (совпадает с coin-cli/go.mod)
ARG GO_VERSION=1.22.0
RUN arch="$(dpkg --print-architecture)" \
    && case "${arch}" in amd64) goarch=amd64;; arm64) goarch=arm64;; *) goarch=amd64;; esac \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz" \
      | tar -C /usr/local -xzf -
ENV PATH="/usr/local/go/bin:${PATH}"

COPY plugins.txt /usr/share/jenkins/ref/plugins.txt
RUN jenkins-plugin-cli --plugin-file /usr/share/jenkins/ref/plugins.txt

COPY casc.yaml /usr/share/jenkins/ref/casc.yaml
COPY casc-jobs.yaml /usr/share/jenkins/ref/casc-jobs.yaml
COPY jobs/ /usr/share/jenkins/ref/jobs/
COPY jenkins-entrypoint.sh /usr/local/bin/jenkins-entrypoint.sh
RUN chmod +x /usr/local/bin/jenkins-entrypoint.sh

ENV JAVA_OPTS="-Djenkins.install.runSetupWizard=false"

USER jenkins
ENTRYPOINT ["/usr/local/bin/jenkins-entrypoint.sh"]
