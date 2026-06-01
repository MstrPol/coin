# Coin managed Dockerfile: Gradle Java runtime image.

FROM gradle:8-jdk{{JAVA_VERSION}} AS builder
WORKDIR /src
COPY settings.gradle build.gradle ./
COPY src ./src
RUN gradle build -x test --no-daemon

FROM eclipse-temurin:{{JAVA_VERSION}}-jre AS runtime
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
ENV COIN_VERSION="${COIN_VERSION}"
RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app
COPY --from=builder --chown=app:app /src/build/libs/*.jar /app/app.jar
USER app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
