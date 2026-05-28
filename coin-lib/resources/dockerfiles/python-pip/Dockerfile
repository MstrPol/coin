# Coin managed Dockerfile: python-pip runtime image.

FROM python:{{PYTHON_VERSION}}-slim-bookworm AS builder
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt ./
RUN python -m venv /app/.venv \
    && . /app/.venv/bin/activate \
    && python -m pip install --upgrade pip \
    && python -m pip install -r requirements.txt

COPY src ./src

FROM python:{{PYTHON_VERSION}}-slim-bookworm AS runtime
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app
COPY --from=builder --chown=app:app /app/.venv /app/.venv
COPY --from=builder --chown=app:app /app/src /app/src
ENV PATH="/app/.venv/bin:$PATH" \
    COIN_VERSION="${COIN_VERSION}" \
    PYTHONPATH="/app/src" \
    PYTHONUNBUFFERED=1
USER app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
