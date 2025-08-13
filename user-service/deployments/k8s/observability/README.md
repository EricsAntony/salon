# Observability Stack for User Service

This directory contains Kubernetes manifests for a comprehensive observability stack including:

- **Grafana Loki + Promtail** for centralized logging
- **Prometheus + Grafana** for metrics collection and visualization
- **OpenTelemetry + Jaeger** for distributed tracing

## Components

### 📊 Metrics (Prometheus + Grafana)
- **Prometheus**: Collects metrics from user-service `/metrics` endpoint
- **Grafana**: Visualizes metrics with pre-configured dashboards
- **Alerts**: Built-in alerting rules for service health monitoring

### 📝 Logs (Loki + Promtail)
- **Loki**: Centralized log aggregation and storage
- **Promtail**: Log shipping agent (DaemonSet) that collects logs from all pods
- **Structured Logging**: JSON format for better querying and filtering

### 🔍 Tracing (Jaeger + OpenTelemetry)
- **Jaeger**: Distributed tracing backend with UI
- **OpenTelemetry**: Instrumentation library integrated in user-service
- **Trace Collection**: Automatic trace collection via OTLP protocol

## Quick Start

### 1. Deploy Observability Stack
```bash
# Make the deployment script executable
chmod +x deploy-observability.sh

# Deploy all components
./deploy-observability.sh
```

### 2. Access UIs
Use kubectl port-forward to access the UIs:

```bash
# Grafana (admin/admin)
kubectl port-forward -n observability svc/grafana 3000:3000

# Prometheus
kubectl port-forward -n observability svc/prometheus 9090:9090

# Jaeger
kubectl port-forward -n observability svc/jaeger 16686:16686

# Loki (API only)
kubectl port-forward -n observability svc/loki 3100:3100
```

### 3. Deploy User Service with Observability
```bash
# Deploy user-service with observability integration
kubectl apply -k ../overlays/dev
```

## Configuration

### User Service Integration

The user-service is automatically configured for observability:

**Metrics**: 
- Prometheus scrapes `/metrics` endpoint
- Custom metrics for OTP operations, user registrations, HTTP requests

**Logs**: 
- Structured JSON logging when `LOG_FORMAT=json`
- Promtail automatically collects logs from all pods
- Logs include service name, namespace, pod name

**Tracing**: 
- OpenTelemetry integration with Jaeger
- Automatic trace collection for HTTP requests
- Environment variables configure Jaeger endpoint

### Environment Variables

The following environment variables are set in the deployment:

```yaml
# Tracing
- name: JAEGER_ENDPOINT
  value: "http://jaeger.observability.svc.cluster.local:14268/api/traces"
- name: OTEL_SERVICE_NAME
  value: "user-service"

# Logging
- name: LOG_FORMAT
  value: "json"
- name: LOG_LEVEL
  value: "info"
```

## Dashboards

### Grafana Dashboards

Pre-configured dashboards include:

1. **User Service Overview**
   - HTTP request rates and response times
   - Error rates and status codes
   - OTP operation metrics
   - User registration/authentication metrics
   - Database query performance

### Prometheus Alerts

Built-in alerting rules:

- **UserServiceDown**: Service is unreachable
- **HighErrorRate**: HTTP 5xx error rate > 10%
- **HighOTPFailureRate**: OTP failure rate > 50%

## Querying

### Loki Log Queries

```logql
# All user-service logs
{job="dev/user-service"}

# Error logs only
{job="dev/user-service"} |= "error"

# OTP-related logs
{job="dev/user-service"} |= "otp"

# Logs from specific pod
{job="dev/user-service", pod="user-service-xxx"}
```

### Prometheus Metrics Queries

```promql
# Request rate
rate(http_requests_total{job="user-service"}[5m])

# Error rate
rate(http_requests_total{job="user-service",status=~"5.."}[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="user-service"}[5m]))

# OTP verification success rate
rate(otp_verifications_total{job="user-service",status="success"}[5m]) / rate(otp_verifications_total{job="user-service"}[5m])
```

## Troubleshooting

### Check Component Status
```bash
kubectl get pods -n observability
kubectl get svc -n observability
```

### View Logs
```bash
# Prometheus logs
kubectl logs -n observability deployment/prometheus

# Grafana logs
kubectl logs -n observability deployment/grafana

# Loki logs
kubectl logs -n observability deployment/loki

# Promtail logs
kubectl logs -n observability daemonset/promtail
```

### Common Issues

1. **Metrics not appearing**: Check Prometheus targets at `http://localhost:9090/targets`
2. **Logs not showing**: Verify Promtail is running on all nodes
3. **Traces missing**: Check Jaeger endpoint configuration in user-service

## Scaling Considerations

For production deployments:

1. **Persistent Storage**: Replace `emptyDir` with persistent volumes
2. **Resource Limits**: Adjust CPU/memory limits based on load
3. **High Availability**: Deploy multiple replicas with proper anti-affinity
4. **Retention Policies**: Configure appropriate data retention periods
5. **Security**: Enable authentication and TLS encryption

## Metrics Reference

### User Service Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | HTTP request duration |
| `otp_requests_total` | Counter | OTP generation requests |
| `otp_verifications_total` | Counter | OTP verification attempts |
| `user_registrations_total` | Counter | User registrations |
| `user_authentications_total` | Counter | User authentications |
| `database_queries_total` | Counter | Database queries |
| `rate_limit_hits_total` | Counter | Rate limit violations |

All metrics include relevant labels for filtering and aggregation.
