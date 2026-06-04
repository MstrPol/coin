# Coin managed Dockerfile: python-pip runtime image (runtime-only, compile в agent).

FROM python:{{PYTHON_VERSION}}-slim-bookworm
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"

RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app

COPY .venv /app/.venv
COPY src /app/src

ENV PATH="/app/.venv/bin:$PATH" \
    COIN_VERSION="${COIN_VERSION}" \
    PYTHONPATH="/app/src" \
    PYTHONUNBUFFERED=1

USER app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
