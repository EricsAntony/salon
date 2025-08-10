# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
WORKDIR /app
COPY user-service/go.mod user-service/go.sum ./
RUN go mod download
COPY user-service/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./cmd

# Install migrate CLI (v4.17.1) in builder and make it available for final image
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget tar \
    && rm -rf /var/lib/apt/lists/* \
    && wget -qO /tmp/migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz \
    && tar -xzf /tmp/migrate.tar.gz -C /usr/local/bin \
    && mv /usr/local/bin/migrate.linux-amd64 /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate \
    && rm -f /tmp/migrate.tar.gz

FROM debian:bookworm-slim
WORKDIR /
COPY --from=builder /app/user-service /usr/local/bin/user-service
COPY user-service/configs/ /configs/
COPY user-service/migrations/ /migrations/
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
EXPOSE 8080
# Create non-root user to run the service
RUN useradd -m -u 10001 appuser \
    && chown -R appuser:appuser /configs /migrations /usr/local/bin/user-service /usr/local/bin/migrate
USER appuser
ENTRYPOINT ["/usr/local/bin/user-service"]
