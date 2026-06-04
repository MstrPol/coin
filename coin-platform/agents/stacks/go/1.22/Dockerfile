# CI agent: go @ 1.22 — GP go-app/v1
FROM golang:1.22-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates docker.io \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace

USER ci
