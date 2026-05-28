FROM maven:3.9-eclipse-temurin-17

WORKDIR /workspace
USER root
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace
USER ci
