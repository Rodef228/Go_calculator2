# Этап сборки
FROM golang:1.23-alpine3.21 AS builder

RUN apk update && apk add --no-cache git gcc g++ libc-dev postgresql-client
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/bin/orchestrator ./cmd/orchestrator/main.go
RUN go build -o /app/bin/agent ./cmd/agent/main.go

# Этап для агентов (минимальный образ)
FROM alpine:3.21 AS agent
COPY --from=builder /app/bin/agent /app/agent
CMD ["/app/agent"]

FROM alpine:3.21 AS orchestrator
RUN apk add --no-cache postgresql-client libc6-compat
COPY --from=builder /app/bin/orchestrator /app/orchestrator
COPY --from=builder /app/migrations /app/migrations
CMD ["/app/orchestrator"]
