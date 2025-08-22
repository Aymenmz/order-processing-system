# Advanced Go Microservices: Order Processing System

[![CI/CD Pipeline](https://github.com/your-org/order-processing-system/actions/workflows/ci-cd.yaml/badge.svg)](https://github.com/your-org/order-processing-system/actions/workflows/ci-cd.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-org/order-processing-system)](https://goreportcard.com/report/github.com/your-org/order-processing-system)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

The Order Processing System is an advanced Go (Golang) microservices project designed to demonstrate enterprise-grade engineering practices. This system showcases the implementation of gRPC-based microservices, comprehensive observability with metrics, logging, and distributed tracing, along with modern DevOps practices including containerization and Kubernetes deployment.

This project serves as a practical learning resource for Cloud DevOps Engineers seeking to master complex engineering concepts and elevate their expertise to top-tier levels. It goes beyond typical web applications or simple CLI tools, focusing on the intricate challenges of distributed systems, service mesh architectures, and production-ready observability.

## Architecture

The system implements a distributed order processing workflow using three core microservices that communicate via gRPC:

- **Order Service**: Orchestrates the order creation process, coordinating with inventory and payment services
- **Inventory Service**: Manages product stock levels and handles reservation/release operations
- **Payment Service**: Processes payment transactions with simulated success/failure scenarios

Each service is instrumented with comprehensive observability features including structured logging with Zap, Prometheus metrics collection, and OpenTelemetry distributed tracing. The entire system is containerized using Docker and can be deployed to Kubernetes with automated scaling and health monitoring.

## Key Features

### Advanced gRPC Implementation
- Protocol Buffers for strongly-typed service contracts
- Unary and streaming RPC support with proper error handling
- Client-side and server-side interceptors for cross-cutting concerns
- Connection pooling and load balancing for high availability

### Comprehensive Observability
- **Structured Logging**: JSON-formatted logs with contextual information including trace IDs, span IDs, and business identifiers
- **Metrics Collection**: Prometheus metrics for both technical (request duration, error rates) and business (orders created, payments processed) KPIs
- **Distributed Tracing**: OpenTelemetry integration with Jaeger for end-to-end request tracing across service boundaries
- **Health Monitoring**: Kubernetes-native health checks with liveness and readiness probes

### Production-Ready DevOps
- **Containerization**: Multi-stage Docker builds optimized for security and size
- **Kubernetes Deployment**: Complete manifests with ConfigMaps, Secrets, Services, and Horizontal Pod Autoscalers
- **CI/CD Pipeline**: GitHub Actions workflow with automated testing, security scanning, and deployment
- **Infrastructure as Code**: Declarative Kubernetes manifests and Docker Compose for reproducible deployments

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Protocol Buffers compiler (`protoc`)
- kubectl (for Kubernetes deployment)

### Local Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-org/order-processing-system.git
   cd order-processing-system
   ```

2. **Install dependencies and generate protobuf code:**
   ```bash
   make setup
   ```

3. **Build all services:**
   ```bash
   make build
   ```

4. **Run tests:**
   ```bash
   make test
   ```

### Docker Compose Deployment

Deploy the entire system with observability stack using Docker Compose:

```bash
make deploy-docker
```

This will start:
- All three microservices with health checks
- Jaeger for distributed tracing (http://localhost:16686)
- Prometheus for metrics collection (http://localhost:9090)
- Grafana for metrics visualization (http://localhost:3000, admin/admin)

### Kubernetes Deployment

Deploy to a Kubernetes cluster:

```bash
make deploy-k8s
```

Access the services using port forwarding:
```bash
kubectl port-forward -n order-processing svc/jaeger 16686:16686
kubectl port-forward -n order-processing svc/prometheus 9090:9090
kubectl port-forward -n order-processing svc/grafana 3000:3000
```

### Testing the System

Run the test client to create sample orders:

```bash
make test-client
```

This will demonstrate the complete order flow: inventory reservation, payment processing, and order completion with full observability traces.

## Project Structure

```
order-processing-system/
├── cmd/                          # Application entry points
│   ├── order-service/           # Order service main applications
│   ├── inventory-service/       # Inventory service main applications
│   ├── payment-service/         # Payment service main applications
│   └── test-client/             # Test client for demonstration
├── pkg/                         # Shared packages and business logic
│   ├── pb/                      # Generated Protocol Buffer code
│   ├── observability/           # Observability utilities and interceptors
│   ├── order/                   # Order service business logic
│   ├── inventory/               # Inventory service business logic
│   └── payment/                 # Payment service business logic
├── api/proto/                   # Protocol Buffer definitions
├── deployments/                 # Deployment configurations
│   ├── docker/                  # Docker Compose files
│   └── kubernetes/              # Kubernetes manifests
├── scripts/                     # Build and deployment scripts
├── .github/workflows/           # CI/CD pipeline definitions
└── docs/                        # Additional documentation
```

## Development Workflow

### Adding New Features

1. **Update Protocol Buffers**: Modify `.proto` files in `api/proto/`
2. **Regenerate Code**: Run `make proto` to update generated code
3. **Implement Business Logic**: Add functionality in respective `pkg/` directories
4. **Add Tests**: Create comprehensive unit and integration tests
5. **Update Documentation**: Keep README and code comments current

### Code Quality

The project enforces high code quality standards:

```bash
make lint          # Run golangci-lint
make fmt           # Format code with gofmt and goimports
make test-coverage # Generate test coverage reports
```

### Local Development

Start services locally for development:

```bash
make dev-run       # Start all services in background
make dev-stop      # Stop all services
```

## Observability Deep Dive

### Metrics

The system exposes comprehensive metrics via Prometheus:

**Technical Metrics:**
- `requests_total`: Total number of gRPC requests by service, method, and status
- `request_duration_seconds`: Request duration histograms for performance monitoring
- `active_connections`: Current number of active gRPC connections

**Business Metrics:**
- `orders_created_total`: Total orders created by status (success/failed)
- `payments_processed_total`: Total payments processed by status
- `inventory_reservations_total`: Total inventory reservations by product and status
- `current_stock`: Current stock levels for all products

### Distributed Tracing

Every request is traced end-to-end using OpenTelemetry:

- **Trace Context Propagation**: Automatic context propagation across gRPC calls
- **Span Attributes**: Rich span metadata including gRPC method names, status codes, and business identifiers
- **Error Recording**: Automatic error capture and span status updates
- **Performance Insights**: Request latency breakdown across service boundaries

### Structured Logging

All services use structured JSON logging with contextual information:

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "caller": "order/service.go:45",
  "msg": "Order created successfully",
  "service": "order-service",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "order_id": "order-123",
  "customer_id": "customer-456",
  "total_amount": 1059.97
}
```

## Deployment Strategies

### Docker Compose (Development)

Ideal for local development and testing:
- Single-command deployment with `make deploy-docker`
- Includes full observability stack
- Automatic service discovery and networking
- Volume persistence for development data

### Kubernetes (Production)

Production-ready deployment with:
- **High Availability**: Multiple replicas with anti-affinity rules
- **Auto Scaling**: Horizontal Pod Autoscalers based on CPU and memory metrics
- **Health Monitoring**: Liveness and readiness probes for all services
- **Resource Management**: CPU and memory requests/limits for optimal scheduling
- **Security**: Non-root containers with minimal attack surface

### CI/CD Pipeline

Automated pipeline with GitHub Actions:

1. **Code Quality**: Linting, testing, and security scanning
2. **Build**: Multi-architecture Docker image builds
3. **Security**: Trivy vulnerability scanning
4. **Deploy**: Automated deployment to staging and production environments
5. **Monitoring**: Post-deployment health checks and rollback capabilities

## Performance Considerations

### gRPC Optimizations

- **Connection Pooling**: Reuse connections across requests
- **Compression**: Automatic gzip compression for large payloads
- **Keep-Alive**: Configurable keep-alive settings for long-lived connections
- **Load Balancing**: Client-side load balancing with health checking

### Resource Management

- **Memory Efficiency**: Optimized Go garbage collection settings
- **CPU Utilization**: Proper goroutine management and context cancellation
- **Network Optimization**: Efficient serialization with Protocol Buffers
- **Storage**: Stateless services with external data persistence

## Security Features

### Container Security

- **Minimal Base Images**: Alpine Linux for reduced attack surface
- **Non-Root Users**: All containers run as non-privileged users
- **Read-Only Filesystems**: Immutable container filesystems where possible
- **Security Scanning**: Automated vulnerability scanning in CI/CD

### Network Security

- **Service Mesh Ready**: Compatible with Istio and other service mesh solutions
- **TLS Encryption**: Support for mutual TLS between services
- **Network Policies**: Kubernetes NetworkPolicies for traffic isolation
- **Secrets Management**: Kubernetes Secrets for sensitive configuration

## Monitoring and Alerting

### Grafana Dashboards

Pre-configured dashboards for:
- **Service Overview**: High-level service health and performance metrics
- **Business Metrics**: Order processing rates, payment success rates, inventory levels
- **Infrastructure**: Resource utilization, container health, and scaling events
- **SLA Monitoring**: Request latency percentiles and error rate tracking

### Alerting Rules

Prometheus alerting rules for:
- **High Error Rates**: Alert when error rates exceed thresholds
- **Latency Degradation**: Alert on P95 latency increases
- **Resource Exhaustion**: Alert on high CPU/memory usage
- **Service Unavailability**: Alert when services become unreachable

## Extending the System

### Adding New Services

1. **Define Protocol Buffers**: Create `.proto` files for new service contracts
2. **Implement Service Logic**: Follow existing patterns in `pkg/` directories
3. **Add Observability**: Integrate metrics, logging, and tracing
4. **Create Deployment Manifests**: Add Docker and Kubernetes configurations
5. **Update CI/CD**: Include new service in build and deployment pipelines

### Integration Patterns

- **Event-Driven Architecture**: Add message brokers (Kafka, RabbitMQ) for asynchronous communication
- **Database Integration**: Add persistent storage with proper connection pooling and migrations
- **External APIs**: Integrate with third-party services using circuit breakers and retries
- **Caching**: Add Redis or similar for performance optimization

## Troubleshooting

### Common Issues

**Service Discovery Problems:**
```bash
kubectl get endpoints -n order-processing  # Check service endpoints
kubectl logs -n order-processing deployment/order-service  # Check service logs
```

**Observability Issues:**
```bash
kubectl port-forward -n order-processing svc/jaeger 16686:16686  # Access Jaeger UI
kubectl port-forward -n order-processing svc/prometheus 9090:9090  # Access Prometheus
```

**Performance Issues:**
- Check Grafana dashboards for resource utilization
- Review distributed traces for bottlenecks
- Analyze Prometheus metrics for error patterns

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
make dev-run
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:
- Code style and standards
- Testing requirements
- Pull request process
- Issue reporting guidelines

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

This project demonstrates advanced patterns and practices from the Go and Cloud Native communities. Special thanks to the maintainers of:
- [gRPC-Go](https://github.com/grpc/grpc-go)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Zap Logging](https://github.com/uber-go/zap)

---

**Built with ❤️ by Manus AI for Cloud DevOps Engineers**

