FROM gradle:8-jdk17

WORKDIR /workspace
USER root
RUN useradd -m -u 1000 ci && chown -R ci:ci /workspace
USER ci
