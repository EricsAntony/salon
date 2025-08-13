#!/bin/bash

# Deploy Observability Stack for User Service
# This script deploys Prometheus, Grafana, Loki, Promtail, and Jaeger

set -e

echo "🚀 Deploying Observability Stack..."

# Create observability namespace
echo "📦 Creating observability namespace..."
kubectl apply -f namespace.yaml

# Deploy Loki (logs storage)
echo "📊 Deploying Loki..."
kubectl apply -f loki.yaml

# Deploy Promtail (log shipper)
echo "🚚 Deploying Promtail..."
kubectl apply -f promtail.yaml

# Deploy Prometheus (metrics)
echo "📈 Deploying Prometheus..."
kubectl apply -f prometheus.yaml

# Deploy Jaeger (tracing)
echo "🔍 Deploying Jaeger..."
kubectl apply -f jaeger.yaml

# Deploy Grafana (visualization)
echo "📊 Deploying Grafana..."
kubectl apply -f grafana.yaml

echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/loki -n observability
kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n observability
kubectl wait --for=condition=available --timeout=300s deployment/jaeger -n observability
kubectl wait --for=condition=available --timeout=300s deployment/grafana -n observability

echo "✅ Observability stack deployed successfully!"
echo ""
echo "🌐 Access URLs (use kubectl port-forward):"
echo "  Grafana:    http://localhost:3000 (admin/admin)"
echo "  Prometheus: http://localhost:9090"
echo "  Jaeger:     http://localhost:16686"
echo "  Loki:       http://localhost:3100"
echo ""
echo "📋 Port-forward commands:"
echo "  kubectl port-forward -n observability svc/grafana 3000:3000"
echo "  kubectl port-forward -n observability svc/prometheus 9090:9090"
echo "  kubectl port-forward -n observability svc/jaeger 16686:16686"
echo "  kubectl port-forward -n observability svc/loki 3100:3100"
