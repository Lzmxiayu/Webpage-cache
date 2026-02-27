FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    chromium \
    fonts-liberation \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/server /app/server

RUN mkdir -p /app/data/screenshots

ENV HTTP_ADDR=:8080
ENV SCREENSHOT_DIR=/app/data/screenshots
ENV SCREENSHOT_BASE_URL=/static/screenshots
ENV SHOT_TIMEOUT_SEC=30
ENV TASK_TIMEOUT_SEC=45
ENV SHUTDOWN_TIMEOUT_SEC=10
ENV WORKER_COUNT=5
ENV MAX_RETRY_COUNT=2
ENV REDIS_QUEUE_KEY=screenshot:tasks
ENV LOG_LEVEL=info

EXPOSE 8080

CMD ["/app/server"]

