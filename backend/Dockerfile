FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/siem-backend .

# Итоговый образ
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/siem-backend /app/siem-backend

EXPOSE 8080

CMD ["/railgun/backend"]