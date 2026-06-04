# CI agent: java-gradle @ 17 — GP java-gradle-app/v1
FROM gradle:8-jdk17

WORKDIR /workspace
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends docker.io \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -u 1000 ci && chown -R ci:ci /workspace

COPY coin-golden-paths /opt/coin-golden-paths
ENV COIN_GOLDEN_PATHS_DIR=/opt/coin-golden-paths
USER ci
