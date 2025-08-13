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

# --- Kubernetes (Kustomize) helpers ---
.PHONY: k8s-ns k8s-apply-dev k8s-apply-stage k8s-apply-prod k8s-delete-dev k8s-delete-stage k8s-delete-prod k8s-secrets-dev k8s-secrets-stage k8s-secrets-prod

K8S_DIR := deployments/k8s

# Create namespaces (idempotent)
k8s-ns:
	kubectl create ns salon-dev --dry-run=client -o yaml | kubectl apply -f -
	kubectl create ns salon-stage --dry-run=client -o yaml | kubectl apply -f -
	kubectl create ns salon-prod --dry-run=client -o yaml | kubectl apply -f -

# Apply overlays
k8s-apply-dev: k8s-ns
	kubectl apply -k $(K8S_DIR)/overlays/dev

k8s-apply-stage: k8s-ns
	kubectl apply -k $(K8S_DIR)/overlays/stage

k8s-apply-prod: k8s-ns
	kubectl apply -k $(K8S_DIR)/overlays/prod

# Delete overlays
k8s-delete-dev:
	kubectl delete -k $(K8S_DIR)/overlays/dev || true

k8s-delete-stage:
	kubectl delete -k $(K8S_DIR)/overlays/stage || true

k8s-delete-prod:
	kubectl delete -k $(K8S_DIR)/overlays/prod || true

# Apply sample secrets (replace with real secrets in CI/CD)
k8s-secrets-dev:
	kubectl apply -f $(K8S_DIR)/overlays/dev/secret.sample.yaml

k8s-secrets-stage:
	kubectl apply -f $(K8S_DIR)/overlays/stage/secret.sample.yaml

k8s-secrets-prod:
	kubectl apply -f $(K8S_DIR)/overlays/prod/secret.sample.yaml
