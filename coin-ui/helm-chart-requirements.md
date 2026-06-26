# Требования к Helm-чарту модуля coin-ui

Данный документ описывает обязательные требования к Helm-чарту для развертывания frontend-модуля `coin-ui`.

## 1. Управление образами и доступами
- **image.repository** и **image.tag**: Обязательная параметризация для выбора деплоимой версии образа.
- **imagePullSecrets**: Обязательный проброс секретов для скачивания из приватных/корпоративных registries.

**Пример `values.yaml`:**
```yaml
image:
  repository: registry.corp.local/coin-ui
  tag: "1.0.0"
  pullSecret: "corp-registry-creds"
```

## 2. Конфигурация и порты
Frontend-приложение компилируется в статику, поэтому большинство конфигураций (например, адрес API) запекаются на этапе сборки. Однако настройки Nginx, обслуживающего `coin-ui`, должны быть правильно сопоставлены в Helm-чарте.

Поскольку `coin-ui` использует Nginx и работает от имени **non-root** пользователя (UID 1000), он слушает порт `8080` (стандартный порт 80 недоступен для non-root). Это важно учесть при настройке Service.

**Фрагмент `values.yaml`:**
```yaml
config:
  port: "8080"
```

**Фрагмент `service.yaml`:**
```yaml
spec:
  ports:
    - port: 80
      targetPort: {{ .Values.config.port }}
```

## 3. Ресурсы и масштабирование
- Обязательно указывать блок `resources`. Поскольку Nginx только раздает статику, лимиты можно сделать минимальными.
- HPA (Horizontal Pod Autoscaler) опционален, но желателен для высоконагруженных сред.

**Пример `values.yaml`:**
```yaml
resources:
  requests:
    cpu: "50m"
    memory: "64Mi"
  limits:
    cpu: "200m"
    memory: "128Mi"
```

## 4. Healthchecks
Эндпоинты проверки жизнеспособности обязательны (простая проверка доступности главной страницы `index.html`).

**Фрагмент `deployment.yaml`:**
```yaml
livenessProbe:
  httpGet:
    path: /
    port: {{ .Values.config.port }}
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /
    port: {{ .Values.config.port }}
  initialDelaySeconds: 5
  periodSeconds: 10
```

## 5. Security Context
Docker-образ настроен на запуск от пользователя с UID 1000. В `deployment.yaml` необходимо это явно отразить:

**Фрагмент `deployment.yaml`:**
```yaml
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  runAsNonRoot: true
```

## 6. Ingress
Маршрутизация к UI является первичной точкой входа пользователей.
- Блок Ingress должен поддерживать настройку доменных имен (`hosts`) и специфичные корпоративные аннотации.

**Пример `values.yaml`:**
```yaml
ingress:
  enabled: true
  hosts:
    - host: coin.corp.local
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: coin-ui-tls
      hosts:
        - coin.corp.local
```
