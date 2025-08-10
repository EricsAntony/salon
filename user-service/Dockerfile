# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
WORKDIR /app
COPY user-service/go.mod user-service/go.sum ./
RUN go mod download
COPY user-service/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./cmd

FROM debian:bookworm-slim
WORKDIR /

# Install psql client
RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

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
