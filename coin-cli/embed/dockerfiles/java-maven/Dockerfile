# Coin managed Dockerfile: Maven Java runtime image.

FROM maven:3.9-eclipse-temurin-{{JAVA_VERSION}} AS builder
WORKDIR /src
COPY pom.xml ./
RUN mvn -B -DskipTests dependency:go-offline
COPY src ./src
RUN mvn -B -DskipTests package

FROM eclipse-temurin:{{JAVA_VERSION}}-jre AS runtime
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
ENV COIN_VERSION="${COIN_VERSION}"
RUN groupadd --gid 1000 app && useradd --uid 1000 --gid app --create-home app
COPY --from=builder --chown=app:app /src/target/*.jar /app/app.jar
USER app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
