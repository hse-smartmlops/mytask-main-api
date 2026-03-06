FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server/main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
RUN cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    echo "Europe/Moscow" > /etc/timezone
WORKDIR /app
COPY --from=builder /app/server ./
CMD ["./server"]