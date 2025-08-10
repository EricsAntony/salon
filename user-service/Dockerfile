# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
WORKDIR /app
COPY user-service/go.mod user-service/go.sum ./
RUN go mod download
COPY user-service/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./cmd

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/user-service /usr/local/bin/user-service
COPY user-service/configs/ /configs/
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/user-service"]
