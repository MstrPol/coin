# Coin managed Dockerfile: Gradle Java runtime image (runtime-only, compile в agent).

FROM eclipse-temurin:{{JAVA_VERSION}}-jre
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
ENV COIN_VERSION="${COIN_VERSION}"

RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app

COPY build/libs/*.jar /app/app.jar

USER app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
