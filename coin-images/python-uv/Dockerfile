# Stack image: Python + uv (pinned)
FROM python:3.13-slim-bookworm

ARG UV_VERSION=0.6.14
ENV UV_LINK_MODE=copy \
    UV_COMPILE_BYTECODE=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir "uv==${UV_VERSION}"

WORKDIR /workspace

# Non-root for typical K8s policies (adjust UID to match cluster)
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace
USER ci
