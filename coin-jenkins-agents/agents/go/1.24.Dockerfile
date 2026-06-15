# CI agent: go @ 1.24 — coin-executor, toolchain go1.24+
FROM golang:1.24-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates docker.io curl unzip \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace

USER ci
