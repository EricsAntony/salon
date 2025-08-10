# syntax=docker/dockerfile:1

# --- Stage 1: Build service + install migrate ---
    FROM golang:1.22 AS builder

    # Install migrate binary (this goes into /go/bin/migrate)
    RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0
    
    WORKDIR /app
    COPY user-service/go.mod user-service/go.sum ./
    RUN go mod download
    COPY user-service/ ./
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./cmd
    
    # --- Stage 2: Final runtime image ---
    FROM debian:bookworm-slim
    
    WORKDIR /
    
    # Install psql client
    RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*
    
    # Copy compiled Go binary
    COPY --from=builder /app/user-service /usr/local/bin/user-service
    
    # Copy configs and migrations
    COPY user-service/configs/ /configs/
    COPY user-service/migrations/ /migrations/
    
    # Copy migrate binary from builder
    COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
    
    EXPOSE 8080
    
    # Create non-root user
    RUN useradd -m -u 10001 appuser \
        && chown -R appuser:appuser /configs /migrations /usr/local/bin/user-service /usr/local/bin/migrate
    
    USER appuser
    
    ENTRYPOINT ["/usr/local/bin/user-service"]
    