# Этап сборки
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Сначала зависимости
COPY go.mod go.sum ./
RUN go mod download

# Потом весь код
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# Этап запуска
FROM alpine:3.20

WORKDIR /app

# Нужны CA-сертификаты, для HTTPS
RUN apk --no-cache add ca-certificates

# Кладём бинарник и миграции
COPY --from=builder /app/server /app/server
COPY migrations /app/migrations

# Порт сервера по умолчанию
ENV SERVER_ADDR=:8080

# Запуск
CMD ["/app/server"]