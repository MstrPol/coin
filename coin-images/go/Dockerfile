FROM golang:1.22-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace

# coin CLI
ARG COIN_CLI_VERSION=0.1.0
ADD ${COIN_CLI_REGISTRY:-artifacts.company.local/coin-cli}/${COIN_CLI_VERSION}/coin_linux_amd64 /usr/local/bin/coin
RUN chmod +x /usr/local/bin/coin
USER ci
