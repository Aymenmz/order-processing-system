#!/bin/bash

# Deployment script for Order Processing System
# This script handles both Docker Compose and Kubernetes deployments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to deploy with Docker Compose
deploy_docker_compose() {
    print_status "Deploying with Docker Compose..."
    
    if ! command_exists docker-compose && ! command_exists docker; then
        print_error "Docker or Docker Compose not found. Please install Docker."
        exit 1
    fi
    
    cd deployments/docker
    
    # Build and start services
    print_status "Building and starting services..."
    if command_exists docker-compose; then
        docker-compose up --build -d
    else
        docker compose up --build -d
    fi
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 30
    
    # Check service health
    print_status "Checking service health..."
    for service in order-service inventory-service payment-service; do
        if command_exists docker-compose; then
            if docker-compose ps | grep -q "$service.*Up"; then
                print_status "$service is running"
            else
                print_warning "$service may not be ready yet"
            fi
        else
            if docker compose ps | grep -q "$service.*running"; then
                print_status "$service is running"
            else
                print_warning "$service may not be ready yet"
            fi
        fi
    done
    
    print_status "Docker Compose deployment completed!"
    print_status "Access points:"
    print_status "  - Jaeger UI: http://localhost:16686"
    print_status "  - Prometheus: http://localhost:9090"
    print_status "  - Grafana: http://localhost:3000 (admin/admin)"
    print_status "  - Order Service gRPC: localhost:50051"
    
    cd ../..
}

# Function to deploy to Kubernetes
deploy_kubernetes() {
    print_status "Deploying to Kubernetes..."
    
    if ! command_exists kubectl; then
        print_error "kubectl not found. Please install kubectl."
        exit 1
    fi
    
    # Check if kubectl is configured
    if ! kubectl cluster-info >/dev/null 2>&1; then
        print_error "kubectl is not configured or cluster is not accessible."
        exit 1
    fi
    
    cd deployments/kubernetes
    
    # Apply namespace
    print_status "Creating namespace..."
    kubectl apply -f namespace.yaml
    
    # Apply observability stack
    print_status "Deploying observability stack..."
    kubectl apply -f observability-stack.yaml
    
    # Wait for observability stack to be ready
    print_status "Waiting for observability stack to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/jaeger -n order-processing
    kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n order-processing
    kubectl wait --for=condition=available --timeout=300s deployment/grafana -n order-processing
    
    # Apply services
    print_status "Deploying microservices..."
    kubectl apply -f payment-service.yaml
    kubectl apply -f inventory-service.yaml
    kubectl apply -f order-service.yaml
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/payment-service -n order-processing
    kubectl wait --for=condition=available --timeout=300s deployment/inventory-service -n order-processing
    kubectl wait --for=condition=available --timeout=300s deployment/order-service -n order-processing
    
    # Get service endpoints
    print_status "Getting service endpoints..."
    kubectl get services -n order-processing
    
    print_status "Kubernetes deployment completed!"
    print_status "Use 'kubectl get pods -n order-processing' to check pod status"
    print_status "Use 'kubectl port-forward -n order-processing svc/jaeger 16686:16686' to access Jaeger UI"
    print_status "Use 'kubectl port-forward -n order-processing svc/prometheus 9090:9090' to access Prometheus"
    print_status "Use 'kubectl port-forward -n order-processing svc/grafana 3000:3000' to access Grafana"
    
    cd ../..
}

# Function to build Docker images
build_images() {
    print_status "Building Docker images..."
    
    if ! command_exists docker; then
        print_error "Docker not found. Please install Docker."
        exit 1
    fi
    
    # Build each service
    for service in order-service inventory-service payment-service; do
        print_status "Building $service..."
        docker build -t "order-processing/$service:latest" -f "Dockerfile.$service" .
    done
    
    print_status "Docker images built successfully!"
}

# Function to run tests
run_tests() {
    print_status "Running tests..."
    
    # Generate protobuf code if needed
    if [ ! -f "pkg/pb/order/order.pb.go" ]; then
        print_status "Generating protobuf code..."
        ./scripts/generate-proto.sh
        
        # Move generated files to correct locations
        mv api/proto/order.pb.go api/proto/order_grpc.pb.go pkg/pb/order/ 2>/dev/null || true
        mv api/proto/inventory.pb.go api/proto/inventory_grpc.pb.go pkg/pb/inventory/ 2>/dev/null || true
        mv api/proto/payment.pb.go api/proto/payment_grpc.pb.go pkg/pb/payment/ 2>/dev/null || true
    fi
    
    # Run Go tests
    go test -v ./...
    
    print_status "Tests completed!"
}

# Function to clean up
cleanup() {
    print_status "Cleaning up..."
    
    case $1 in
        docker)
            cd deployments/docker
            if command_exists docker-compose; then
                docker-compose down -v
            else
                docker compose down -v
            fi
            cd ../..
            ;;
        kubernetes)
            kubectl delete namespace order-processing --ignore-not-found=true
            ;;
        *)
            print_warning "Specify cleanup target: docker or kubernetes"
            ;;
    esac
    
    print_status "Cleanup completed!"
}

# Main script logic
case $1 in
    docker)
        build_images
        deploy_docker_compose
        ;;
    kubernetes|k8s)
        build_images
        deploy_kubernetes
        ;;
    build)
        build_images
        ;;
    test)
        run_tests
        ;;
    cleanup)
        cleanup $2
        ;;
    *)
        echo "Usage: $0 {docker|kubernetes|k8s|build|test|cleanup}"
        echo ""
        echo "Commands:"
        echo "  docker      - Deploy using Docker Compose"
        echo "  kubernetes  - Deploy to Kubernetes cluster"
        echo "  k8s         - Alias for kubernetes"
        echo "  build       - Build Docker images only"
        echo "  test        - Run tests"
        echo "  cleanup     - Clean up deployment (specify 'docker' or 'kubernetes')"
        echo ""
        echo "Examples:"
        echo "  $0 docker"
        echo "  $0 kubernetes"
        echo "  $0 cleanup docker"
        exit 1
        ;;
esac

