# Coin managed Dockerfile: Go runtime image (runtime-only, compile в agent).
# Не копируйте в репозитории сервисов.

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
ENV COIN_VERSION="${COIN_VERSION}"
COPY dist/app /app/app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
