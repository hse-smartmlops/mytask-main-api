# Этап сборки
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o server cmd/server/main.go

# Минимальный образ
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata bash

WORKDIR /app

COPY --from=builder /app/server ./
COPY --from=builder /app/.env ./

CMD ["./server"]
