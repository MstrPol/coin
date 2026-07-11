# syntax=docker/dockerfile:1.8
ARG REGISTRY=""
ARG NPM_REGISTRY="https://registry.npmjs.org/"
ARG NODE_TLS_REJECT_UNAUTHORIZED=0

FROM ${REGISTRY}node:22-alpine AS build
WORKDIR /app

ENV NODE_TLS_REJECT_UNAUTHORIZED=${NODE_TLS_REJECT_UNAUTHORIZED}

# Кэшируем npm-зависимости для ускорения сборки
COPY package.json package-lock.json* ./
# Используем секрет npm_token для доступа к приватным пакетам
RUN --mount=type=cache,target=/root/.npm \
    --mount=type=secret,id=npm_token \
    npm config set strict-ssl false && \
    npm config set registry ${NPM_REGISTRY} && \
    if [ -f /run/secrets/npm_token ]; then npm config set "${NPM_REGISTRY#http*:}:_authToken" "$(cat /run/secrets/npm_token)"; fi && \
    npm ci 2>/dev/null || npm install

COPY . .
ENV VITE_API_BASE=/api

# Сборка проекта (используем кэш npm)
RUN --mount=type=cache,target=/root/.npm \
    npm run build

# Финальный образ
FROM ${REGISTRY}nginx:1.27-alpine

# Настраиваем Nginx для работы от non-root пользователя (id 1000)
# Создаем пользователя app и выдаем ему права на директории Nginx
RUN addgroup -g 1000 app && adduser -u 1000 -G app -D app && \
    chown -R app:app /var/cache/nginx /var/log/nginx /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && chown -R app:app /var/run/nginx.pid

# Копируем конфигурацию и собранную статику
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

# Устанавливаем права на скопированные файлы
RUN chown -R app:app /usr/share/nginx/html /etc/nginx/conf.d

USER app
EXPOSE 8080
