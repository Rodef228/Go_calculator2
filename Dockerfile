# =============================
# Этап сборки
# =============================
FROM golang:1.23-alpine3.21 AS builder

# Установка необходимых пакетов
RUN apk update && \
    apk add --no-cache git gcc g++ libc-dev postgresql-client

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка бинарников
RUN CGO_ENABLED=0 go build -o /app/bin/orchestrator ./cmd/orchestrator/main.go
RUN CGO_ENABLED=0 go build -o /app/bin/agent ./cmd/agent/main.go

# =============================
# Этап для агентов
# =============================
FROM alpine:3.21 AS agent

RUN apk add --no-cache libc6-compat

COPY --from=builder /app/bin/agent /app/agent
WORKDIR /app

CMD ["/app/agent"]

# =============================
# Этап для оркестратора
# =============================
FROM alpine:3.21 AS orchestrator

RUN apk add --no-cache postgresql-client libc6-compat ca-certificates

COPY --from=builder /app/bin/orchestrator /app/orchestrator
COPY --from=builder /app/migrations /app/migrations

WORKDIR /app

EXPOSE 8080

CMD ["/app/orchestrator"]