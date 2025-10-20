FROM golang:1.24.5-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY .env ./ 
CMD ["./main"]