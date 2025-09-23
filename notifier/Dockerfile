FROM golang:1.23-alpine AS builder

# Install necessary packages for building
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set the working directory
WORKDIR /build

# Copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o notifier .

# Final image
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add \
    ca-certificates \
    sqlite \
    dcron \
    tzdata \
    && rm -rf /var/cache/apk/*

# Set the timezone (can be changed if needed)
ENV TZ=Europe/Kiev

# Create a user for security
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# Create working directories
WORKDIR /app
RUN mkdir -p /app/data /var/log/cron && \
    chown -R appuser:appgroup /app /var/log/cron

# Copy the binary from the builder stage
COPY --from=builder /build/notifier /app/
COPY --from=builder /build/.env /app/
COPY --from=builder /build/crontab /tmp/

# Create crontab for the appuser
RUN crontab -u appuser /tmp/crontab && \
    rm /tmp/crontab

# Create a startup script
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "Starting notifier cron service..."' >> /app/start.sh && \
    echo 'echo "Cron will run notifier every 5 minutes"' >> /app/start.sh && \
    echo '# Start cron as root, then switch to appuser for the actual process' >> /app/start.sh && \
    echo 'crond -f -d 8' >> /app/start.sh && \
    chmod +x /app/start.sh /app/notifier && \
    chown appuser:appgroup /app/start.sh

# DO NOT switch user - crond must be run as root
# USER appuser

# Expose a volume for data
VOLUME ["/app/data"]

# Start command
CMD ["/app/start.sh"]
