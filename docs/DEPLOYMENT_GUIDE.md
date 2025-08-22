# Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Docker Compose Deployment](#docker-compose-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Production Deployment](#production-deployment)
6. [Monitoring and Observability Setup](#monitoring-and-observability-setup)
7. [CI/CD Pipeline Configuration](#cicd-pipeline-configuration)
8. [Troubleshooting](#troubleshooting)
9. [Maintenance and Updates](#maintenance-and-updates)

## Prerequisites

### System Requirements

**Development Environment:**
- Operating System: Linux (Ubuntu 20.04+), macOS (10.15+), or Windows 10+ with WSL2
- RAM: Minimum 8GB, Recommended 16GB
- CPU: Minimum 4 cores, Recommended 8 cores
- Disk Space: Minimum 10GB free space
- Network: Stable internet connection for downloading dependencies

**Software Dependencies:**
- Go 1.21 or later
- Docker 20.10+ and Docker Compose 2.0+
- kubectl 1.25+ (for Kubernetes deployment)
- Protocol Buffers compiler (protoc) 3.15+
- Git 2.30+
- Make utility

### Installation Instructions

**Go Installation:**
```bash
# Linux/macOS
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Verify installation
go version
```

**Docker Installation:**
```bash
# Ubuntu
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker-compose --version
```

**Protocol Buffers Installation:**
```bash
# Ubuntu
sudo apt-get update
sudo apt-get install -y protobuf-compiler

# macOS
brew install protobuf

# Verify installation
protoc --version
```

**kubectl Installation:**
```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# macOS
brew install kubectl

# Verify installation
kubectl version --client
```

## Local Development Setup

### Repository Setup

**Clone and Initialize:**
```bash
# Clone the repository
git clone https://github.com/your-org/order-processing-system.git
cd order-processing-system

# Install Go dependencies and tools
make setup

# Verify setup
make test
```

**Development Tools Installation:**
```bash
# Install additional development tools
make install-tools

# This installs:
# - protoc-gen-go (Protocol Buffers Go plugin)
# - protoc-gen-go-grpc (gRPC Go plugin)
# - golangci-lint (Go linter)
```

### Environment Configuration

**Environment Variables:**
Create a `.env` file for local development:
```bash
# Service Configuration
ORDER_SERVICE_PORT=50051
INVENTORY_SERVICE_PORT=50052
PAYMENT_SERVICE_PORT=50053

# Metrics Ports
ORDER_METRICS_PORT=8080
INVENTORY_METRICS_PORT=8081
PAYMENT_METRICS_PORT=8082

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
LOG_LEVEL=info

# Development Settings
ENVIRONMENT=development
DEBUG_MODE=false
```

**IDE Configuration:**
For Visual Studio Code, create `.vscode/settings.json`:
```json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.formatTool": "goimports",
    "go.testFlags": ["-v"],
    "files.associations": {
        "*.proto": "proto3"
    }
}
```

### Local Development Workflow

**Start Services for Development:**
```bash
# Build all services
make build

# Start services in development mode
make dev-run

# Services will be available at:
# - Order Service: localhost:50051 (gRPC), localhost:8080 (metrics)
# - Inventory Service: localhost:50052 (gRPC), localhost:8081 (metrics)
# - Payment Service: localhost:50053 (gRPC), localhost:8082 (metrics)
```

**Testing the Setup:**
```bash
# Run the test client
make test-client

# Expected output:
# Creating order...
# Order created successfully: order-abc123
# Total amount: $2059.97
# Status: ORDER_STATUS_PROCESSING
# ...
```

**Stop Development Services:**
```bash
make dev-stop
```

## Docker Compose Deployment

### Quick Start

**Deploy Complete System:**
```bash
# Deploy with observability stack
make deploy-docker

# This starts:
# - All microservices
# - Jaeger (tracing)
# - Prometheus (metrics)
# - Grafana (visualization)
```

**Access Points:**
- Jaeger UI: http://localhost:16686
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Order Service: localhost:50051

### Custom Configuration

**Override Configuration:**
Create `deployments/docker/docker-compose.override.yml`:
```yaml
version: '3.8'

services:
  order-service:
    environment:
      - LOG_LEVEL=debug
      - CUSTOM_CONFIG=value
    ports:
      - "50051:50051"
      - "8080:8080"
      - "9000:9000"  # Additional port

  prometheus:
    volumes:
      - ./custom-prometheus.yml:/etc/prometheus/prometheus.yml
```

**Environment-Specific Deployment:**
```bash
# Development environment
docker-compose -f docker-compose.yaml -f docker-compose.dev.yml up -d

# Staging environment
docker-compose -f docker-compose.yaml -f docker-compose.staging.yml up -d
```

### Monitoring Docker Deployment

**Check Service Status:**
```bash
# View running services
docker-compose ps

# View service logs
docker-compose logs order-service
docker-compose logs -f --tail=100 inventory-service

# View all logs with timestamps
docker-compose logs -t
```

**Health Checks:**
```bash
# Check service health endpoints
curl http://localhost:8080/health  # Order Service
curl http://localhost:8081/health  # Inventory Service
curl http://localhost:8082/health  # Payment Service

# Check metrics endpoints
curl http://localhost:8080/metrics  # Order Service metrics
```

### Cleanup

**Stop and Remove:**
```bash
# Stop services
docker-compose down

# Remove volumes and networks
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## Kubernetes Deployment

### Cluster Prerequisites

**Kubernetes Cluster Requirements:**
- Kubernetes version 1.25+
- Minimum 3 worker nodes
- 4 CPU cores and 8GB RAM per node
- Container runtime (Docker or containerd)
- Ingress controller (nginx, traefik, etc.)
- Storage class for persistent volumes

**Local Kubernetes Options:**
```bash
# minikube
minikube start --cpus=4 --memory=8192 --kubernetes-version=v1.28.0

# kind
kind create cluster --config=deployments/kubernetes/kind-config.yaml

# Docker Desktop (enable Kubernetes in settings)
```

### Deployment Process

**Deploy to Kubernetes:**
```bash
# Deploy complete system
make deploy-k8s

# Manual deployment steps:
kubectl apply -f deployments/kubernetes/namespace.yaml
kubectl apply -f deployments/kubernetes/observability-stack.yaml
kubectl apply -f deployments/kubernetes/payment-service.yaml
kubectl apply -f deployments/kubernetes/inventory-service.yaml
kubectl apply -f deployments/kubernetes/order-service.yaml
```

**Verify Deployment:**
```bash
# Check namespace
kubectl get namespaces

# Check all resources
kubectl get all -n order-processing

# Check pod status
kubectl get pods -n order-processing -w

# Check service endpoints
kubectl get endpoints -n order-processing
```

### Access Services

**Port Forwarding:**
```bash
# Access Jaeger UI
kubectl port-forward -n order-processing svc/jaeger 16686:16686

# Access Prometheus
kubectl port-forward -n order-processing svc/prometheus 9090:9090

# Access Grafana
kubectl port-forward -n order-processing svc/grafana 3000:3000

# Access Order Service
kubectl port-forward -n order-processing svc/order-service 50051:50051
```

**Ingress Configuration:**
Create `deployments/kubernetes/ingress.yaml`:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: order-processing-ingress
  namespace: order-processing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: order-processing.local
    http:
      paths:
      - path: /jaeger
        pathType: Prefix
        backend:
          service:
            name: jaeger
            port:
              number: 16686
      - path: /prometheus
        pathType: Prefix
        backend:
          service:
            name: prometheus
            port:
              number: 9090
      - path: /grafana
        pathType: Prefix
        backend:
          service:
            name: grafana
            port:
              number: 3000
```

### Scaling and Updates

**Manual Scaling:**
```bash
# Scale order service
kubectl scale deployment order-service -n order-processing --replicas=5

# Scale all services
kubectl scale deployment --all -n order-processing --replicas=3
```

**Rolling Updates:**
```bash
# Update service image
kubectl set image deployment/order-service -n order-processing \
  order-service=order-processing/order-service:v2.0.0

# Check rollout status
kubectl rollout status deployment/order-service -n order-processing

# Rollback if needed
kubectl rollout undo deployment/order-service -n order-processing
```

**Auto-Scaling Configuration:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: order-service-hpa
  namespace: order-processing
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: order-service
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Production Deployment

### Production Readiness Checklist

**Infrastructure Requirements:**
- [ ] Multi-zone Kubernetes cluster with at least 3 nodes per zone
- [ ] Persistent storage with backup and replication
- [ ] Load balancer with SSL termination
- [ ] DNS configuration for service discovery
- [ ] Monitoring and alerting infrastructure
- [ ] Log aggregation and retention policies
- [ ] Backup and disaster recovery procedures

**Security Configuration:**
- [ ] Network policies for service isolation
- [ ] RBAC configuration for service accounts
- [ ] Secrets management for sensitive data
- [ ] Image vulnerability scanning
- [ ] Runtime security monitoring
- [ ] Compliance with security standards (SOC2, PCI-DSS, etc.)

**Performance Optimization:**
- [ ] Resource requests and limits tuned for workload
- [ ] Horizontal and vertical auto-scaling configured
- [ ] Database connection pooling and optimization
- [ ] CDN configuration for static assets
- [ ] Caching strategies implemented
- [ ] Performance testing completed

### Production Configuration

**Resource Limits and Requests:**
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

**Production Environment Variables:**
```yaml
env:
- name: ENVIRONMENT
  value: "production"
- name: LOG_LEVEL
  value: "info"
- name: METRICS_ENABLED
  value: "true"
- name: TRACING_SAMPLE_RATE
  value: "0.1"  # 10% sampling in production
- name: DATABASE_URL
  valueFrom:
    secretKeyRef:
      name: database-secret
      key: url
```

**Health Check Configuration:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### Multi-Environment Strategy

**Environment Separation:**
```bash
# Development
kubectl apply -f deployments/kubernetes/overlays/development/

# Staging
kubectl apply -f deployments/kubernetes/overlays/staging/

# Production
kubectl apply -f deployments/kubernetes/overlays/production/
```

**Kustomization Example:**
```yaml
# deployments/kubernetes/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../../base

patchesStrategicMerge:
- production-config.yaml

replicas:
- name: order-service
  count: 5
- name: inventory-service
  count: 3
- name: payment-service
  count: 3

images:
- name: order-processing/order-service
  newTag: v1.2.3
- name: order-processing/inventory-service
  newTag: v1.2.3
- name: order-processing/payment-service
  newTag: v1.2.3
```

## Monitoring and Observability Setup

### Prometheus Configuration

**Custom Metrics Configuration:**
```yaml
# prometheus-config.yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: 'order-processing-services'
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - order-processing
    relabel_configs:
    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
```

**Alert Rules:**
```yaml
# alert_rules.yml
groups:
- name: order-processing-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(requests_total{status="error"}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} for service {{ $labels.service }}"

  - alert: HighLatency
    expr: histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High latency detected"
      description: "95th percentile latency is {{ $value }}s for {{ $labels.service }}"
```

### Grafana Dashboards

**Service Overview Dashboard:**
```json
{
  "dashboard": {
    "title": "Order Processing System Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(requests_total[5m])) by (service)",
            "legendFormat": "{{ service }}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(requests_total{status=\"error\"}[5m])) by (service)",
            "legendFormat": "{{ service }} errors"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(request_duration_seconds_bucket[5m])) by (le, service))",
            "legendFormat": "{{ service }} p95"
          }
        ]
      }
    ]
  }
}
```

### Log Management

**Centralized Logging with ELK Stack:**
```yaml
# filebeat-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: filebeat-config
  namespace: order-processing
data:
  filebeat.yml: |
    filebeat.inputs:
    - type: container
      paths:
        - /var/log/containers/*order-processing*.log
      processors:
      - add_kubernetes_metadata:
          host: ${NODE_NAME}
          matchers:
          - logs_path:
              logs_path: "/var/log/containers/"

    output.elasticsearch:
      hosts: ["elasticsearch:9200"]
      index: "order-processing-%{+yyyy.MM.dd}"

    setup.template.name: "order-processing"
    setup.template.pattern: "order-processing-*"
```

## CI/CD Pipeline Configuration

### GitHub Actions Setup

**Repository Secrets:**
Configure the following secrets in your GitHub repository:
- `DOCKER_REGISTRY_URL`: Container registry URL
- `DOCKER_REGISTRY_USERNAME`: Registry username
- `DOCKER_REGISTRY_PASSWORD`: Registry password
- `KUBE_CONFIG`: Base64 encoded kubeconfig file
- `SLACK_WEBHOOK_URL`: Slack notification webhook (optional)

**Environment-Specific Deployments:**
```yaml
# .github/workflows/deploy-staging.yml
name: Deploy to Staging

on:
  push:
    branches: [ develop ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: staging
    steps:
    - uses: actions/checkout@v4
    
    - name: Deploy to Staging
      run: |
        echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig
        kubectl apply -f deployments/kubernetes/overlays/staging/
        kubectl rollout status deployment/order-service -n order-processing-staging
```

### GitLab CI/CD

**GitLab CI Configuration:**
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

test:
  stage: test
  image: golang:1.21
  script:
    - make setup
    - make test
    - make lint

build:
  stage: build
  image: docker:20.10.16
  services:
    - docker:20.10.16-dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - make docker-build
    - make docker-push

deploy-staging:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl config use-context $KUBE_CONTEXT_STAGING
    - kubectl apply -f deployments/kubernetes/overlays/staging/
  only:
    - develop

deploy-production:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl config use-context $KUBE_CONTEXT_PRODUCTION
    - kubectl apply -f deployments/kubernetes/overlays/production/
  only:
    - main
  when: manual
```

## Troubleshooting

### Common Issues and Solutions

**Service Discovery Problems:**
```bash
# Check DNS resolution
kubectl exec -it deployment/order-service -n order-processing -- nslookup inventory-service

# Check service endpoints
kubectl get endpoints -n order-processing

# Check service labels and selectors
kubectl describe service inventory-service -n order-processing
kubectl get pods -n order-processing --show-labels
```

**Performance Issues:**
```bash
# Check resource utilization
kubectl top pods -n order-processing
kubectl top nodes

# Check HPA status
kubectl get hpa -n order-processing
kubectl describe hpa order-service-hpa -n order-processing

# Check metrics
curl http://localhost:8080/metrics | grep -E "(cpu|memory|requests)"
```

**Networking Issues:**
```bash
# Check network policies
kubectl get networkpolicies -n order-processing

# Test connectivity between services
kubectl exec -it deployment/order-service -n order-processing -- \
  grpc_cli call inventory-service:50052 inventory.InventoryService.GetProductStock \
  'product_id: "product-1"'
```

**Observability Issues:**
```bash
# Check Jaeger connectivity
kubectl logs deployment/order-service -n order-processing | grep jaeger

# Check Prometheus targets
kubectl port-forward -n order-processing svc/prometheus 9090:9090
# Visit http://localhost:9090/targets

# Check metrics endpoints
kubectl exec -it deployment/order-service -n order-processing -- \
  curl http://localhost:8080/metrics
```

### Debug Mode

**Enable Debug Logging:**
```yaml
# Add to deployment environment variables
- name: LOG_LEVEL
  value: "debug"
- name: JAEGER_SAMPLER_TYPE
  value: "const"
- name: JAEGER_SAMPLER_PARAM
  value: "1"  # 100% sampling for debugging
```

**Debug Container:**
```bash
# Run debug container in the same namespace
kubectl run debug --image=nicolaka/netshoot -it --rm -n order-processing

# Inside the debug container:
nslookup order-service
curl http://order-service:8080/health
grpc_cli call order-service:50051 order.OrderService.GetOrder 'order_id: "test"'
```

### Log Analysis

**Common Log Patterns:**
```bash
# Filter error logs
kubectl logs deployment/order-service -n order-processing | grep '"level":"error"'

# Filter by trace ID
kubectl logs deployment/order-service -n order-processing | grep "trace_id:abc123"

# Monitor logs in real-time
kubectl logs -f deployment/order-service -n order-processing --tail=100
```

## Maintenance and Updates

### Regular Maintenance Tasks

**Weekly Tasks:**
- Review and rotate logs
- Check resource utilization trends
- Update security patches
- Review monitoring alerts and thresholds
- Backup configuration and data

**Monthly Tasks:**
- Update base images and dependencies
- Review and update resource limits
- Performance testing and optimization
- Security vulnerability scanning
- Disaster recovery testing

**Quarterly Tasks:**
- Major version updates
- Architecture review and optimization
- Capacity planning and scaling review
- Security audit and compliance review
- Documentation updates

### Update Procedures

**Rolling Updates:**
```bash
# Update single service
kubectl set image deployment/order-service -n order-processing \
  order-service=order-processing/order-service:v1.2.4

# Monitor rollout
kubectl rollout status deployment/order-service -n order-processing

# Rollback if issues occur
kubectl rollout undo deployment/order-service -n order-processing
```

**Blue-Green Deployment:**
```bash
# Deploy to green environment
kubectl apply -f deployments/kubernetes/overlays/production-green/

# Test green environment
kubectl port-forward -n order-processing-green svc/order-service 50051:50051

# Switch traffic to green
kubectl patch service order-service -n order-processing \
  -p '{"spec":{"selector":{"version":"green"}}}'

# Remove blue environment after validation
kubectl delete namespace order-processing-blue
```

### Backup and Recovery

**Configuration Backup:**
```bash
# Backup all Kubernetes resources
kubectl get all -n order-processing -o yaml > backup-$(date +%Y%m%d).yaml

# Backup specific resources
kubectl get configmaps -n order-processing -o yaml > configmaps-backup.yaml
kubectl get secrets -n order-processing -o yaml > secrets-backup.yaml
```

**Disaster Recovery:**
```bash
# Restore from backup
kubectl apply -f backup-20240115.yaml

# Verify restoration
kubectl get pods -n order-processing
make test-client
```

This comprehensive deployment guide provides detailed instructions for deploying and maintaining the Order Processing System across different environments, from local development to production Kubernetes clusters. The guide emphasizes best practices for security, monitoring, and operational excellence.

