# CI agent: java-maven @ 17 — GP java-maven-app/v1
FROM maven:3.9-eclipse-temurin-17

WORKDIR /workspace
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends docker.io curl unzip \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -u 1000 ci && chown -R ci:ci /workspace

USER ci
