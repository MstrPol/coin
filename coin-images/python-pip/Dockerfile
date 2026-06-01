FROM python:3.13-slim-bookworm

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir pytest build twine

WORKDIR /workspace
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace

# coin CLI
ARG COIN_CLI_VERSION=0.1.0
ADD ${COIN_CLI_REGISTRY:-artifacts.company.local/coin-cli}/${COIN_CLI_VERSION}/coin_linux_amd64 /usr/local/bin/coin
RUN chmod +x /usr/local/bin/coin
USER ci
