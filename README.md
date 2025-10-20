# Emplacc API

Сервис предоставляет HTTP API (Echo), gRPC клиент для MCP и интеграции с Keycloak, MinIO, Prometheus.

## Архитектура

```
cmd/
  server/                     # Точка входа HTTP приложения

api/
  v1/
    dto/                      # Запросы, ответы, swagger-обёртки
    handlers/                 # Echo-хендлеры
    middleware/               # HTTP middleware (trace/meta, валидация, rate-limit)
    presenter/                # Преобразование доменных сущностей в DTO
    router.go                 # Регистрация маршрутов

internal/
  app/
    container.go              # Сборка зависимостей (DI контейнер)
    runtime.go                # Старт приложения, инициализация сервисов
    ports/
      ports.go                # Общие типы (Pagination, ObjectStorage, TokenValidator)
      entity.go               # Контракты DTO для сервисов
      repository.go           # Интерфейсы репозиториев
      service.go              # Интерфейсы сервисов
      mcp.go                  # Контракты MCP (репозиторий/сервис, стримы)
  config/                     # Конфигурация, ENV загрузка
  domain/
    models/                   # GORM-модели
    ...                       # Доменные ошибки и вспомогательные типы
  migrations/                 # Автомиграции БД
  repo/
    auth/                     # Интеграция с Keycloak
    mcp/                      # gRPC MCP репозиторий
    minio/                    # MinIO адаптер (опционально)
    pg/                       # PostgreSQL репозитории
  service/
    mcp.go                    # Сервис поверх MCP (sync + stream)
    ...                       # Прочие доменные сервисы
  transport/
    grpc/
      v1/                     # gRPC клиент MCP
    http/
      middleware/             # Общие Echo middleware (rate-limiter, trace meta)

pkg/
  excel/                      # Экспорт отчётов в XLSX
  logger/                     # Настройка slog
  metrics/                    # Prometheus middleware / handler
  pb/v1/                      # Сгенерированные protobuf типы
  swagger/                    # swagger.Doc конфигурация + сгенерированные схемы
  tracing/                    # OpenTelemetry конфигурация
  utils/                      # Утилиты (значения, преобразования)

proto/
  emplacc/
    v1/
      mcp.proto               # Исходный proto-контракт MCP

docs/                         # (если используются дополнительные документы)
.env.example / .env.test      # Конфигурация окружений
```

- `internal/app/container.go` собирает зависимости и принимает `ports.MCPService` извне.
- `internal/app/ports` разделён на тематические файлы (`entity`, `repository`, `service`, `ports`, `mcp`).
- `internal/service/mcp.go` реализует синхронный и потоковый API поверх gRPC.

## Генерация артефактов

```bash
protoc -I proto \
  --go_out=.. \
  --go-grpc_out=.. \
  proto/emplacc/v1/mcp.proto
```

```bash
swag init -g cmd/server/main.go -o pkg/swagger/docs --parseDependency --parseInternal
```

## Потоковая работа с MCP

- `POST /v1/tasks/{id}/improve-report` — синхронное улучшение текста.
- `GET /v1/tasks/{id}/improve-report/ws` — потоковая передача событий по WebSocket. После апгрейда клиент отправляет JSON вида:

```json
{
  "user_text": "...",
  "meta": {"user_id": "..."},
  "timeout_ms": 60000,
  "content_type": "text/markdown"
}
```

далее получает события вида:

```
- {"type":"chunk","chunk":{"data":"...","index":1}}
```

- События содержат тип (`status`, `chunk`, `final`, `error`) и полезные данные. Поток завершается, когда приходит `final` или `error`.
- На фронтенде достаточно стандартного API: `const ws = new WebSocket("/v1/tasks/:id/improve-report/ws"); ws.onmessage = (ev) => { ... }`. При необходимости SSE можно реализовать через промежуточный прокси.

## Переменные окружения

Группы переменных повторяют `.env.example`:

### Application
- `APP_NAME` — имя приложения.
- `APP_VERSION` — версия.
- `APP_ENV` — окружение (`development`, `staging`, `production`).
- `APP_TIMEZONE` — часовой пояс (например, `Europe/Moscow`).

### Server
- `SERVER_HOST`, `SERVER_PORT` — адрес и порт HTTP сервера.
- `SERVER_READ_TIMEOUT`, `SERVER_WRITE_TIMEOUT`, `SERVER_SHUTDOWN_TIMEOUT` — тайм-ауты.
- `SERVER_ALLOWED_ORIGINS`, `SERVER_ALLOWED_METHODS`, `SERVER_ALLOWED_HEADERS` — CORS.
- `SERVER_ALLOW_CREDENTIALS` — разрешить передачу cookies/authorization в CORS.

### Database
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` — параметры подключения PostgreSQL.
- `DB_SSLMODE`, `DB_TIMEZONE` — режим SSL и часовой пояс.
- `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME` — пул соединений.
- `DB_AUTO_MIGRATE` — выполнять ли автмиграции.

### Logger
- `LOG_LEVEL`, `LOG_FORMAT` — уровень и формат логов (`text`/`json`).
- `LOG_FILE` — путь к файлу (пусто = stdout).
- `LOG_MAX_SIZE_MB`, `LOG_MAX_BACKUPS`, `LOG_MAX_AGE_DAYS`, `LOG_COMPRESS` — ротация логов.

### Swagger
- `SWAGGER_ENABLED` — включить Swagger UI.
- `SWAGGER_TITLE`, `SWAGGER_DESCRIPTION`, `SWAGGER_VERSION`, `SWAGGER_HOST`, `SWAGGER_BASE_PATH`, `SWAGGER_SCHEMES` — метаданные документации.

### Keycloak
- `KEYCLOAK_URL`, `KEYCLOAK_REALM` — адрес и realm.
- `KEYCLOAK_CLIENT_ID`, `KEYCLOAK_CLIENT_SECRET` — клиент для API.
- `KEYCLOAK_BACKEND_CLIENT_ID`, `KEYCLOAK_BACKEND_CLIENT_SECRET` — client для токен-обмена.
- `KEYCLOAK_TOKEN_EXCHANGE_ENABLED` — включить exchange.
- `KEYCLOAK_ADMIN_USER`, `KEYCLOAK_ADMIN_PASSWORD` — администратор dev-стенда.

### MinIO
- `MINIO_ENDPOINT`, `MINIO_USE_SSL`, `MINIO_REGION` — параметры подключения.
- `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY` — креды.
- `MINIO_AVATAR_BUCKET`, `MINIO_REPORT_BUCKET` — бакеты.

### Rate Limits
- `RATE_LIMIT_USER_REQUESTS`, `RATE_LIMIT_USER_WINDOW` — лимит на пользователя.
- `RATE_LIMIT_IP_REQUESTS`, `RATE_LIMIT_IP_WINDOW` — лимит на IP.

### Tracing
- `TRACING_ENABLED` — включить OTel.
- `TRACING_PROVIDER`, `TRACING_ENDPOINT` — настройки провайдера (OTLP, Jaeger и т.д.).
- `TRACING_SERVICE_NAME` — имя сервиса.
- `TRACING_SAMPLE_RATE` — доля выборки (0–1).

### MCP (gRPC)
- `MCP_GRPC_ADDRESS` — адрес MCP сервера (`host:port`).
- `MCP_GRPC_TIMEOUT` — тайм-аут запросов.
- `MCP_GRPC_USE_TLS` — использовать TLS.

### Pagination
- `PAGINATION_DEFAULT_LIMIT`, `PAGINATION_MAX_LIMIT` — границы пагинации.

## Миграции и таблицы БД

- `internal/migrations` содержит миграции GORM. Автозапуск включается флагом `DB_AUTO_MIGRATE=true`.
- Основные таблицы (см. `internal/domain/models`): `users`, `roles`, `user_role`, `projects`, `boards`, `statuses`, `tasks`, `teams`, `team_member`, `project_team`, `subscriptions`, `attendance`, `daily_reports`, `completed_work`, `help_request`, `tomorrow_plans`, `report_problem`, `forum_messages`, `problems`.
- Репозитории в `internal/repo/pg` выполняют preload связей (см. методы `List*`, `Get*`).

## Контракты, сервисы и репозитории

- `internal/app/ports` — интерфейсы (`repository.go`, `service.go`) и DTO (`entity.go`, `mcp.go`).
- `internal/repo/pg`, `internal/repo/minio`, `internal/repo/mcp` — реализации портов.
- `internal/service` — бизнес-логика: CRUD, LLM (`mcp.go`), управление MinIO и т.д.

## Валидация запросов

- `api/v1/middleware/validator.go` — `BindAndValidate`, унифицированные ошибки (`RespondValidationError`), валидация `ImproveTaskReport` и остальных DTO.
- Дополнительные проверки выполняются в хендлерах (например, перед стартом WebSocket).

## Аутентификация и авторизация

- Keycloak (обязателен): репозиторий `internal/repo/auth`, сервис `internal/service/auth.go`.
- Middleware `v1middleware.KeycloakAuth` подключается через `container.V1Deps().AuthMiddleware`.

## HTTP-слой

- `api/v1/handlers` — REST/WS-хендлеры (по сущностям).
- `api/v1/presenter` — преобразование доменных моделей в DTO.
- `api/v1/middleware` — TraceMeta, rate-limit, валидация.
- `api/v1/router.go` — регистрация маршрутов (84 маршрута, включая LLM-эндпоинты).

## Запуск и сборка через Docker

```bash
# сборка образа
docker build -t emplacc-api .

# запуск
docker run --rm -p 8081:8081 --env-file .env emplacc-api
```

Для комплексного окружения (Keycloak, Postgres, MinIO, MCP) используйте Docker Compose:

```yaml
services:
  api:
    build: .
    env_file: .env
    ports:
      - "8081:8081"
    depends_on:
      - postgres
      - keycloak
      - minio
      - mcp
  # ... остальные сервисы
```

```bash
docker compose up --build
```

## Дополнительно

- `.env.example` содержит значения по умолчанию с комментариями.
- `.env.test` — минимальная конфигурация для `docker compose`.
- Prometheus метрики доступны по `/metrics` (см. `pkg/metrics`).
- MCP клиент создаётся автоматически, если указан `MCP_GRPC_ADDRESS` (при ошибке соединения API продолжит работу без LLM-функционала).
- MinIO используется для работы с аватарами и экспортом отчётов; при пустой конфигурации функционал gracefully отключается.
- Keycloak обязателен: при отсутствии конфигурации запуск завершится с ошибкой.
