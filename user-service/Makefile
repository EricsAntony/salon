APP := user-service

.PHONY: build run test tidy docker compose-up compose-down compose-logs compose-ps compose-restart migrate-up-compose migrate-down-compose

build:
	go build -o bin/$(APP) ./cmd

run:
	go run ./cmd

test:
	go test ./...

tidy:
	go mod tidy

docker:
	docker build -t $(APP):local .

# Migrations (requires golang-migrate installed)
MIGRATE_DB_URL ?= postgres://user:password@localhost:5432/salon_users?sslmode=disable
migrate-up:
	migrate -database $(MIGRATE_DB_URL) -path ./migrations up
migrate-down:
	migrate -database $(MIGRATE_DB_URL) -path ./migrations down 1
migrate-new:
	@read -p "Name: " name; migrate create -ext sql -dir ./migrations -seq $$name

# Docker Compose helpers
compose-up:
	docker compose up -d --build

compose-down:
	docker compose down -v

compose-logs:
	docker compose logs -f user-service

compose-ps:
	docker compose ps

compose-restart:
	docker compose restart user-service

# Run migrations inside compose using migrate container
migrate-up-compose:
	docker compose run --rm migrate -path /migrations -database postgres://user:password@db:5432/salon_users?sslmode=disable up

migrate-down-compose:
	docker compose run --rm migrate -path /migrations -database postgres://user:password@db:5432/salon_users?sslmode=disable down 1
