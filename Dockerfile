# Этап сборки
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates bash tzdata

WORKDIR /app

# (Опционально) Очистка кэша модулей, если нужно быть уверенным в свежести
# RUN go clean -modcache # Раскомментируйте, если хотите принудительно очистить кэш модулей перед загрузкой

COPY go.mod go.sum ./
# Загрузка зависимостей (использует кэш слоёв Docker, если go.mod/go.sum не менялись)
RUN go mod download && go mod verify

COPY . ./

# Очистка кэша сборки Go перед сборкой для обеспечения "чистой" сборки
RUN go clean -cache && \
    go build -v -o server cmd/server/main.go # -v для подробности

# Минимальный образ
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata bash

# Скопировать зону в /etc/localtime
RUN cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    echo "Europe/Moscow" > /etc/timezone

WORKDIR /root/

# Копирование бинарного файла из builder stage
COPY --from=builder /app/server ./server
# Копирование .env файла, если он нужен напрямую в корень или в /app
COPY --from=builder /app/.env ./ # Или WORKDIR /app выше и копирование в ./

CMD ["./server"]