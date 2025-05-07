# Этап сборки
FROM golang:1.23-alpine3.21 AS builder

# Устанавливаем базовые пакеты для сборки приложения
RUN apk update && apk add --no-cache ca-certificates git gcc g++ libc-dev binutils

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Копируем весь код приложения
COPY . .

# Сборка исполняемых файлов
RUN go build -o /app/bin/orchestrator ./cmd/orchestrator/main.go
RUN go build -o /app/bin/agent ./cmd/agent/main.go

# Этап выполнения
FROM alpine:3.21 AS runner

# Устанавливаем необходимые пакеты
RUN apk update && apk add --no-cache ca-certificates libc6-compat openssh bash sqlite

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем исполняемые файлы и веб-ресурсы из этапа сборки
COPY --from=builder /app/bin/orchestrator /app/orchestrator
COPY --from=builder /app/bin/agent /app/agent

# Открываем порт для приложения
EXPOSE 8080

# Указываем команду для запуска приложения
CMD ["/app/orchestrator"]
