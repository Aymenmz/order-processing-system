# Makefile for Order Processing System

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
ORDER_BINARY=bin/order-service
INVENTORY_BINARY=bin/inventory-service
PAYMENT_BINARY=bin/payment-service
TEST_CLIENT_BINARY=bin/test-client

# Docker parameters
DOCKER_REGISTRY=order-processing
DOCKER_TAG=latest

# Protobuf parameters
PROTO_PATH=api/proto
PROTO_FILES=$(wildcard $(PROTO_PATH)/*.proto)

.PHONY: all build clean test deps proto docker-build docker-push deploy-docker deploy-k8s help

# Default target
all: deps proto build test

# Build all binaries
build: $(ORDER_BINARY) $(INVENTORY_BINARY) $(PAYMENT_BINARY) $(TEST_CLIENT_BINARY)

$(ORDER_BINARY):
	$(GOBUILD) -o $(ORDER_BINARY) ./cmd/order-service/main_enhanced.go

$(INVENTORY_BINARY):
	$(GOBUILD) -o $(INVENTORY_BINARY) ./cmd/inventory-service/main_enhanced.go

$(PAYMENT_BINARY):
	$(GOBUILD) -o $(PAYMENT_BINARY) ./cmd/payment-service/main_enhanced.go

$(TEST_CLIENT_BINARY):
	$(GOBUILD) -o $(TEST_CLIENT_BINARY) ./cmd/test-client

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -rf pkg/pb/*/*.pb.go

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	./scripts/generate-proto.sh
	@echo "Moving generated files to correct locations..."
	@mv api/proto/order.pb.go api/proto/order_grpc.pb.go pkg/pb/order/ 2>/dev/null || true
	@mv api/proto/inventory.pb.go api/proto/inventory_grpc.pb.go pkg/pb/inventory/ 2>/dev/null || true
	@mv api/proto/payment.pb.go api/proto/payment_grpc.pb.go pkg/pb/payment/ 2>/dev/null || true

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -s -w .
	goimports -w .

# Build Docker images
docker-build:
	docker build -t $(DOCKER_REGISTRY)/order-service:$(DOCKER_TAG) -f Dockerfile.order-service .
	docker build -t $(DOCKER_REGISTRY)/inventory-service:$(DOCKER_TAG) -f Dockerfile.inventory-service .
	docker build -t $(DOCKER_REGISTRY)/payment-service:$(DOCKER_TAG) -f Dockerfile.payment-service .

# Push Docker images
docker-push: docker-build
	docker push $(DOCKER_REGISTRY)/order-service:$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/inventory-service:$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/payment-service:$(DOCKER_TAG)

# Deploy with Docker Compose
deploy-docker:
	./scripts/deploy.sh docker

# Deploy to Kubernetes
deploy-k8s:
	./scripts/deploy.sh kubernetes

# Run services locally for development
dev-run:
	@echo "Starting services in development mode..."
	@echo "Starting Payment Service on port 50053..."
	@./bin/payment-service-enhanced &
	@sleep 2
	@echo "Starting Inventory Service on port 50052..."
	@./bin/inventory-service-enhanced &
	@sleep 2
	@echo "Starting Order Service on port 50051..."
	@./bin/order-service-enhanced &
	@echo "All services started. Use 'make dev-stop' to stop them."

# Stop development services
dev-stop:
	@echo "Stopping development services..."
	@pkill -f "payment-service-enhanced" || true
	@pkill -f "inventory-service-enhanced" || true
	@pkill -f "order-service-enhanced" || true
	@echo "Services stopped."

# Run test client
test-client: $(TEST_CLIENT_BINARY)
	./$(TEST_CLIENT_BINARY)

# Install development tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Setup development environment
setup: install-tools deps proto build
	@echo "Development environment setup complete!"

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Build everything (deps, proto, build, test)"
	@echo "  build         - Build all binaries"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  proto         - Generate protobuf code"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-push   - Build and push Docker images"
	@echo "  deploy-docker - Deploy with Docker Compose"
	@echo "  deploy-k8s    - Deploy to Kubernetes"
	@echo "  dev-run       - Run services locally for development"
	@echo "  dev-stop      - Stop development services"
	@echo "  test-client   - Run test client"
	@echo "  install-tools - Install development tools"
	@echo "  setup         - Setup development environment"
	@echo "  help          - Show this help message"

