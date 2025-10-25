# Assembly stage
FROM golang:1.23 AS builder
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy sources
COPY . .

# Build the binary
RUN go build -o bot ./main.go
#RUN go build -o bot .

# Launch stage
FROM debian:bookworm-slim
WORKDIR /app

# Install SQLite (if need sqlite3 CLI)
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Europe/Kyiv

# Copy binary from builder
COPY --from=builder /app/bot .

# Copy all configurations files and dirs
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/config ./config
COPY --from=builder /app/prompts ./prompts

#  Create dir for DB
RUN mkdir -p /app/data

# ENV for token
ENV TELEGRAM_TOKEN=""

CMD ["./bot"]