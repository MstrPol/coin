# Требования к Helm-чарту модуля coin-api

Данный документ описывает обязательные требования к Helm-чарту для развертывания модуля `coin-api` (согласно стандартам и требованиям к переносимости).

## 1. Управление образами и доступами
- **image.repository** и **image.tag**: Параметризация образа для деплоя конкретных версий.
- **imagePullSecrets**: Проброс секретов для приватных registries.

**Пример `values.yaml`:**
```yaml
image:
  repository: registry.corp.local/coin-api
  tag: "1.0.0"
  pullSecret: "corp-registry-creds"
```

## 2. Конфигурация и переменные окружения (Environment Variables)

Модуль `coin-api` взаимодействует с базой данных PostgreSQL (для хранения метаданных) и Nexus (в качестве кэша манифестов). Для его корректной работы требуются следующие переменные окружения.

### Необходимые переменные
1. **База данных (PostgreSQL)**
   - `DB_HOST` / `DB_PORT`: Адрес и порт инстанса PostgreSQL.
   - `DB_NAME`: Имя базы данных (например, `coin_metadata`).
   - `DB_USER` / `DB_PASSWORD`: Учетные данные (пароль должен браться строго из Secret).
2. **Интеграция с Nexus**
   - `NEXUS_URL`: Базовый адрес Nexus.
   - `NEXUS_USER` / `NEXUS_PASSWORD`: Креды для авторизации в Nexus (из Secret).
3. **API Сервер**
   - `COIN_OPENAPI_PATH`: Путь к OpenAPI спецификации (зашито в Dockerfile: `/usr/share/coin/openapi/v1.yaml`).
   - `PORT`: Порт, который слушает приложение (по умолчанию `8090`).

### Пример реализации конфигурации

**Фрагмент `values.yaml`:**
```yaml
config:
  dbHost: "coin-pg-cluster.db.svc.cluster.local"
  dbPort: "5432"
  dbName: "coin"
  nexusUrl: "http://nexus.nexus.svc.cluster.local:8081"
  port: "8090"

secretName: "coin-api-secrets" # Имя секрета, содержащего DB_PASSWORD и NEXUS_PASSWORD
```

**Фрагмент `deployment.yaml`:**
```yaml
env:
  - name: DB_HOST
    value: {{ .Values.config.dbHost | quote }}
  - name: DB_PORT
    value: {{ .Values.config.dbPort | quote }}
  - name: DB_NAME
    value: {{ .Values.config.dbName | quote }}
  - name: NEXUS_URL
    value: {{ .Values.config.nexusUrl | quote }}
  - name: PORT
    value: {{ .Values.config.port | quote }}
  - name: COIN_OPENAPI_PATH
    value: "/usr/share/coin/openapi/v1.yaml"
  # Секретные значения
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: {{ .Values.secretName }}
        key: db-password
  - name: NEXUS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: {{ .Values.secretName }}
        key: nexus-password
```

## 3. Ресурсы и масштабирование
- Обязательно указывать блок `resources`.
- Рекомендуется включить поддержку HPA для автомасштабирования API при нагрузке.

**Пример `values.yaml`:**
```yaml
resources:
  requests:
    cpu: "200m"
    memory: "256Mi"
  limits:
    cpu: "1000m"
    memory: "1Gi"

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 80
```

## 4. Healthchecks
Эндпоинты проверки жизнеспособности обязательны.

**Фрагмент `deployment.yaml`:**
```yaml
livenessProbe:
  httpGet:
    path: /health # Или другой актуальный эндпоинт в coin-api
    port: {{ .Values.config.port }}
  initialDelaySeconds: 10
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /health
    port: {{ .Values.config.port }}
  initialDelaySeconds: 5
  periodSeconds: 10
```

## 5. Security Context
Docker-образ настроен на запуск от пользователя с UID 1000. В `deployment.yaml` необходимо это отразить:

**Фрагмент `deployment.yaml`:**
```yaml
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  runAsNonRoot: true
```
