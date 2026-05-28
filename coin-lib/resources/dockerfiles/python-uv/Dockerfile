# Coin managed Dockerfile: python-uv runtime image.
# Не копируйте этот файл в репозитории сервисов. Coin генерирует Dockerfile в CI.

FROM ghcr.io/astral-sh/uv:python{{PYTHON_VERSION}}-bookworm-slim AS builder
WORKDIR /app
ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy

COPY pyproject.toml uv.lock* ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-install-project --no-dev

COPY . .
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

FROM python:{{PYTHON_VERSION}}-slim-bookworm AS runtime
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"

RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app

COPY --from=builder --chown=app:app /app/.venv /app/.venv
COPY --from=builder --chown=app:app /app/src /app/src

ENV PATH="/app/.venv/bin:$PATH" \
    COIN_VERSION="${COIN_VERSION}" \
    PYTHONUNBUFFERED=1

USER app
EXPOSE {{APP_PORT}}

CMD {{APP_CMD}}
