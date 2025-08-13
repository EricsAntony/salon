# Local Testing with Docker Desktop Kubernetes

This guide will help you test the complete observability stack locally using Docker Desktop with Kubernetes enabled.

## Prerequisites

### 1. Enable Kubernetes in Docker Desktop
1. Open Docker Desktop
2. Go to Settings → Kubernetes
3. Check "Enable Kubernetes"
4. Click "Apply & Restart"
5. Wait for Kubernetes to start (green indicator)

### 2. Verify Kubernetes is Running
```bash
kubectl cluster-info
kubectl get nodes
```

### 3. Build User Service Docker Image
```bash
# Build the user-service image
docker build -t user-service:latest .

# Verify the image exists
docker images | grep user-service
```

## Step-by-Step Deployment

### Step 1: Deploy Observability Stack
```bash
# Deploy the complete observability stack
make observability-deploy

# Or deploy manually
kubectl apply -f deployments/k8s/observability/namespace.yaml
kubectl apply -f deployments/k8s/observability/loki.yaml
kubectl apply -f deployments/k8s/observability/promtail.yaml
kubectl apply -f deployments/k8s/observability/prometheus.yaml
kubectl apply -f deployments/k8s/observability/jaeger.yaml
kubectl apply -f deployments/k8s/observability/grafana.yaml
```

### Step 2: Wait for Observability Components
```bash
# Check status
kubectl get pods -n observability

# Wait for all pods to be ready
kubectl wait --for=condition=ready pod --all -n observability --timeout=300s
```

### Step 3: Deploy User Service
```bash
# Create dev namespace and deploy user-service
kubectl create namespace dev
kubectl apply -f deployments/k8s/local/secrets.yaml -n dev
kubectl apply -f deployments/k8s/local/deployment.yaml -n dev

# Check user-service status
kubectl get pods -n dev
```

### Step 4: Access UIs via Port Forwarding

Open separate terminal windows for each:

```bash
# Terminal 1: Grafana (admin/admin)
kubectl port-forward -n observability svc/grafana 3000:3000

# Terminal 2: Prometheus
kubectl port-forward -n observability svc/prometheus 9090:9090

# Terminal 3: Jaeger
kubectl port-forward -n observability svc/jaeger 16686:16686

# Terminal 4: User Service API
kubectl port-forward -n dev svc/user-service 8080:8080
```

## Testing the Stack

### 1. Access the UIs
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **User Service**: http://localhost:8080

### 2. Generate Test Traffic
```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Metrics endpoint
curl http://localhost:8080/metrics

# Generate some API traffic
curl -X POST http://localhost:8080/api/v1/otp/request \
  -H "Content-Type: application/json" \
  -d '{"phone": "+1234567890"}'
```

### 3. Verify Observability Data

**Prometheus Metrics:**
1. Go to http://localhost:9090
2. Try these queries:
   ```promql
   # Request rate
   rate(http_requests_total[5m])
   
   # Error rate
   rate(http_requests_total{status=~"5.."}[5m])
   
   # Response time
   histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
   ```

**Grafana Dashboards:**
1. Go to http://localhost:3000 (admin/admin)
2. Navigate to Dashboards → User Service Overview
3. You should see metrics from your test requests

**Jaeger Traces:**
1. Go to http://localhost:16686
2. Select "user-service" from the service dropdown
3. Click "Find Traces" to see distributed traces

**Loki Logs:**
1. In Grafana, go to Explore
2. Select "Loki" as data source
3. Try these queries:
   ```logql
   {job="dev/user-service"}
   {job="dev/user-service"} |= "error"
   {job="dev/user-service"} |= "otp"
   ```

## Troubleshooting

### Common Issues

**1. Pods Not Starting**
```bash
# Check pod status
kubectl get pods -n observability
kubectl describe pod <pod-name> -n observability

# Check logs
kubectl logs <pod-name> -n observability
```

**2. User Service Not Connecting to Database**
```bash
# Check if you have a local PostgreSQL running
# Or update the secret with a test database URL
kubectl get secret user-service-secrets -n dev -o yaml
```

**3. Metrics Not Appearing**
```bash
# Check if user-service is exposing metrics
curl http://localhost:8080/metrics

# Check Prometheus targets
# Go to http://localhost:9090/targets
```

**4. Logs Not Appearing**
```bash
# Check Promtail is running
kubectl get pods -n observability | grep promtail

# Check Promtail logs
kubectl logs daemonset/promtail -n observability
```

### Resource Requirements

For local testing, ensure Docker Desktop has sufficient resources:
- **Memory**: At least 4GB allocated to Docker
- **CPU**: At least 2 cores
- **Disk**: At least 10GB free space

### Cleanup

```bash
# Clean up everything
make clean-dev
make observability-clean

# Or manually
kubectl delete namespace dev
kubectl delete namespace observability
```

## Local Development Workflow

### 1. Code → Build → Deploy → Test
```bash
# Make code changes
# Build new image
docker build -t user-service:latest .

# Restart deployment to pick up new image
kubectl rollout restart deployment/user-service -n dev

# Test changes
curl http://localhost:8080/health
```

### 2. View Real-time Logs
```bash
# Follow user-service logs
kubectl logs -f deployment/user-service -n dev

# Or view in Grafana Loki
# Go to Grafana → Explore → Loki → {job="dev/user-service"}
```

### 3. Monitor Performance
- Watch metrics in Grafana dashboards
- Check traces in Jaeger for performance bottlenecks
- Monitor resource usage: `kubectl top pods -n dev`

## Production Simulation

To simulate production-like conditions locally:

```bash
# Deploy to multiple environments
make deploy-stage-full
make deploy-prod-full

# Generate load across environments
# Use different ports for each environment
kubectl port-forward -n stage svc/user-service 8081:8080 &
kubectl port-forward -n prod svc/user-service 8082:8080 &
```

This setup gives you a complete local testing environment that mirrors production observability capabilities!
