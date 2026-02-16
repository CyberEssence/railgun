# Сборка
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Разрешаем Go автоматически скачивать нужную версию (1.25.6)
ENV GOTOOLCHAIN=auto

RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения с явным указанием версии
RUN go version && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo 'dev') -X main.buildTime=$(date -u +%Y%m%d%H%M%S)" \
    -o railgun-core ./cmd/server

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/railgun-core .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 -S railgun && \
    adduser -u 1000 -S railgun -G railgun && \
    chown -R railgun:railgun /app

USER railgun:railgun

EXPOSE 8080
EXPOSE 2112

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./railgun-core"]