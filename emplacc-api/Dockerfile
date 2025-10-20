# Стадия сборки
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates bash tzdata

# Копируем go.mod и go.sum в /app/emplacc-api
WORKDIR /app/emplacc-api
COPY emplacc-api/go.mod emplacc-api/go.sum ./

# Загружаем зависимости
RUN go mod download && go mod verify

# Копируем ВЕСЬ код из emplacc-api в текущую директорию
COPY emplacc-api/ ./

# Собираем сервер из корня модуля
RUN go clean -cache && \
    go build -v -o server ./cmd/server/main.go

# Финальный образ
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata bash
RUN cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    echo "Europe/Moscow" > /etc/timezone

WORKDIR /app
COPY --from=builder /app/emplacc-api/server ./
COPY .env ./ 
CMD ["./server"]