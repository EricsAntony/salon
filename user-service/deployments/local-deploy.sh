#!/usr/bin/env bash
set -euo pipefail

# ==============================
# CONFIG
# ==============================
GRAFANA_PORT=3000
PROMETHEUS_PORT=9090
JAEGER_PORT=16686
USER_SERVICE_PORT=8080

# ==============================
# FUNCTIONS
# ==============================
check_command() {
    local cmd="$1"
    local install_hint="$2"

    if ! command -v "$cmd" &>/dev/null; then
        echo "❌ $cmd not found."
        echo "   Install with: $install_hint"
        exit 1
    else
        echo "✅ $cmd is installed."
    fi
}

wait_for_namespace() {
    local namespace="$1"
    echo "⏳ Waiting for all pods in namespace '$namespace' to be ready..."
    kubectl wait --for=condition=ready pod --all -n "$namespace" --timeout=300s
    echo "✅ All pods in '$namespace' are ready."
}

# ==============================
# PRECHECKS
# ==============================
echo "🔍 Checking prerequisites..."
check_command kubectl "https://kubernetes.io/docs/tasks/tools/"
check_command docker "https://docs.docker.com/get-docker/"
check_command make "sudo apt install make  # or brew install make"

echo "⚙️ Verifying Kubernetes cluster is running..."
kubectl cluster-info >/dev/null
kubectl get nodes

# ==============================
# BUILD USER-SERVICE
# ==============================
echo "📦 Building user-service Docker image..."
docker build -t user-service:latest .

# ==============================
# DEPLOY OBSERVABILITY STACK
# ==============================
echo "🚀 Deploying observability stack..."
make observability-deploy || {
    kubectl apply -f deployments/k8s/observability/namespace.yaml
    kubectl apply -f deployments/k8s/observability/loki.yaml
    kubectl apply -f deployments/k8s/observability/promtail.yaml
    kubectl apply -f deployments/k8s/observability/prometheus.yaml
    kubectl apply -f deployments/k8s/observability/jaeger.yaml
    kubectl apply -f deployments/k8s/observability/grafana.yaml
}

wait_for_namespace observability

# ==============================
# DEPLOY USER-SERVICE
# ==============================
echo "🚀 Deploying user-service..."
make namespace-dev || kubectl create namespace dev
make secrets-dev
make deploy-dev

wait_for_namespace dev

# ==============================
# PORT FORWARDING
# ==============================
echo "🔌 Starting port-forwarding..."
echo "Grafana: http://localhost:${GRAFANA_PORT} (admin/admin)"
kubectl port-forward -n observability svc/grafana ${GRAFANA_PORT}:3000 &
echo "Prometheus: http://localhost:${PROMETHEUS_PORT}"
kubectl port-forward -n observability svc/prometheus ${PROMETHEUS_PORT}:9090 &
echo "Jaeger: http://localhost:${JAEGER_PORT}"
kubectl port-forward -n observability svc/jaeger ${JAEGER_PORT}:16686 &
echo "User Service API: http://localhost:${USER_SERVICE_PORT}"
kubectl port-forward -n dev svc/user-service ${USER_SERVICE_PORT}:8080 &

# ==============================
# TEST ENDPOINTS
# ==============================
sleep 5
echo "🔍 Running health checks..."
curl -s http://localhost:${USER_SERVICE_PORT}/health || echo "Health check failed!"
curl -s http://localhost:${USER_SERVICE_PORT}/metrics | head -n 5

echo "✅ Deployment completed! Access Grafana, Prometheus, Jaeger, and User Service via the above URLs."
wait
