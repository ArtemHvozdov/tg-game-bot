# Этап сборки
FROM golang:1.23 AS builder
WORKDIR /app

# Скопировать go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Скопировать исходники
COPY . .

# Собрать бинарник
RUN go build -o bot ./main.go

# Этап запуска
FROM debian:bookworm-slim
WORKDIR /app

# Установим SQLite (если нужно sqlite3 CLI)
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Europe/Kyiv

# Скопировать бинарь из builder
COPY --from=builder /app/bot .

# Скопировать все необходимые конфигурационные файлы и директории
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/config ./config
COPY --from=builder /app/prompts ./prompts

# Создать папку для базы (на всякий случай)
RUN mkdir -p /app/data

# ENV для токена (значение передаётся при запуске)
ENV TELEGRAM_TOKEN=""

CMD ["./bot"]