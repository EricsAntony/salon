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
MIGRATE_DB_URL ?= postgres://user:password@localhost:5432/salon?sslmode=disable
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
	docker compose run --rm migrate -path /migrations -database postgres://user:password@db:5432/salon?sslmode=disable up

migrate-down-compose:
	docker compose run --rm migrate -path /migrations -database postgres://user:password@db:5432/salon?sslmode=disable down 1

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

# Clean up
clean-dev:
	kubectl delete -k deployments/k8s/overlays/dev --ignore-not-found=true

clean-stage:
	kubectl delete -k deployments/k8s/overlays/stage --ignore-not-found=true

clean-prod:
	kubectl delete -k deployments/k8s/overlays/prod --ignore-not-found=true

# Apply sample secrets (for testing - use proper secret management in production)
secrets-dev:
	kubectl apply -f deployments/k8s/overlays/dev/sample-secrets.yaml

secrets-stage:
	kubectl apply -f deployments/k8s/overlays/stage/sample-secrets.yaml

secrets-prod:
	kubectl apply -f deployments/k8s/overlays/prod/sample-secrets.yaml

# Observability Stack
.PHONY: observability-deploy observability-clean observability-status observability-port-forward

observability-deploy:
	@echo "🚀 Deploying observability stack..."
	kubectl apply -f deployments/k8s/observability/namespace.yaml
	kubectl apply -f deployments/k8s/observability/loki.yaml
	kubectl apply -f deployments/k8s/observability/promtail.yaml
	kubectl apply -f deployments/k8s/observability/prometheus.yaml
	kubectl apply -f deployments/k8s/observability/jaeger.yaml
	kubectl apply -f deployments/k8s/observability/grafana.yaml
	@echo "⏳ Waiting for deployments to be ready..."
	kubectl wait --for=condition=available --timeout=300s deployment/loki -n observability
	kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n observability
	kubectl wait --for=condition=available --timeout=300s deployment/jaeger -n observability
	kubectl wait --for=condition=available --timeout=300s deployment/grafana -n observability
	@echo "✅ Observability stack deployed successfully!"

observability-clean:
	@echo "🧹 Cleaning up observability stack..."
	kubectl delete -f deployments/k8s/observability/ --ignore-not-found=true
	kubectl delete namespace observability --ignore-not-found=true

observability-status:
	@echo "📊 Observability Stack Status:"
	kubectl get pods -n observability
	@echo ""
	kubectl get svc -n observability

observability-port-forward:
	@echo "🌐 Setting up port forwards for observability UIs..."
	@echo "Access URLs:"
	@echo "  Grafana:    http://localhost:3000 (admin/admin)"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Jaeger:     http://localhost:16686"
	@echo ""
	@echo "Run these commands in separate terminals:"
	@echo "  kubectl port-forward -n observability svc/grafana 3000:3000"
	@echo "  kubectl port-forward -n observability svc/prometheus 9090:9090"
	@echo "  kubectl port-forward -n observability svc/jaeger 16686:16686"

# Full deployment with observability
deploy-dev-full: namespace-dev secrets-dev observability-deploy deploy-dev
	@echo "🎉 Full development environment deployed with observability!"

deploy-stage-full: namespace-stage secrets-stage observability-deploy deploy-stage
	@echo "🎉 Full staging environment deployed with observability!"

deploy-prod-full: namespace-prod secrets-prod observability-deploy deploy-prod
	@echo "🎉 Full production environment deployed with observability!"
