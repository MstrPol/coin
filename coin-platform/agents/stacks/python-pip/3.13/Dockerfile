# CI agent: python-pip @ 3.13 — GP python-pip-app/v1
FROM python:3.13-slim-bookworm

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates docker.io curl unzip \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir pytest build twine

WORKDIR /workspace
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace

USER ci
