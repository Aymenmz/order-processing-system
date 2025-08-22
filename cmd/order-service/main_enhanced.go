package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/order-processing-system/pkg/observability"
	"github.com/your-org/order-processing-system/pkg/order"
	orderpb "github.com/your-org/order-processing-system/pkg/pb/order"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// orderServiceServer implements the gRPC OrderService interface with observability
type orderServiceServer struct {
	orderpb.UnimplementedOrderServiceServer
	service order.Service
	logger  *zap.Logger
}

// CreateOrder handles order creation requests with enhanced logging and metrics
func (s *orderServiceServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// Add business context to logger
	contextLogger := observability.LoggerWithCustomerID(
		observability.LoggerWithTraceContext(ctx, s.logger),
		req.CustomerId,
	)

	contextLogger.Info("Processing CreateOrder request",
		zap.String("customer_id", req.CustomerId),
		zap.Int("items_count", len(req.Items)))

	order, err := s.service.CreateOrder(ctx, req.CustomerId, req.Items)
	if err != nil {
		contextLogger.Error("Failed to create order", zap.Error(err))
		observability.OrdersCreated.WithLabelValues("failed").Inc()
		return nil, err
	}

	// Record successful order creation
	observability.OrdersCreated.WithLabelValues("success").Inc()

	contextLogger.Info("Order created successfully",
		zap.String("order_id", order.Id),
		zap.Float64("total_amount", order.TotalAmount))

	return &orderpb.CreateOrderResponse{Order: order}, nil
}

// GetOrder handles order retrieval requests
func (s *orderServiceServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	contextLogger := observability.LoggerWithOrderID(
		observability.LoggerWithTraceContext(ctx, s.logger),
		req.OrderId,
	)

	contextLogger.Debug("Processing GetOrder request")

	order, err := s.service.GetOrder(ctx, req.OrderId)
	if err != nil {
		contextLogger.Error("Failed to get order", zap.Error(err))
		return nil, err
	}

	contextLogger.Debug("Order retrieved successfully")
	return &orderpb.GetOrderResponse{Order: order}, nil
}

// UpdateOrderStatus handles order status update requests
func (s *orderServiceServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	contextLogger := observability.LoggerWithOrderID(
		observability.LoggerWithTraceContext(ctx, s.logger),
		req.OrderId,
	)

	contextLogger.Info("Processing UpdateOrderStatus request",
		zap.String("new_status", req.NewStatus.String()))

	order, err := s.service.UpdateOrderStatus(ctx, req.OrderId, req.NewStatus)
	if err != nil {
		contextLogger.Error("Failed to update order status", zap.Error(err))
		return nil, err
	}

	contextLogger.Info("Order status updated successfully",
		zap.String("status", order.Status.String()))

	return &orderpb.UpdateOrderStatusResponse{Order: order}, nil
}

func main() {
	serviceName := "order-service"

	// Initialize structured logger
	logger, err := observability.NewLogger(serviceName, zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Order Service with observability")

	// Initialize metrics
	observability.InitMetrics()

	// Initialize tracing
	jaegerEndpoint := getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	cleanup, err := observability.InitTracing(serviceName, jaegerEndpoint, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracing", zap.Error(err))
	}
	defer cleanup()

	// Get service addresses from environment variables
	inventoryAddr := getEnv("INVENTORY_SERVICE_ADDR", "localhost:50052")
	paymentAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50053")
	port := getEnv("PORT", "50051")
	metricsPort := getEnv("METRICS_PORT", "8080")

	// Connect to inventory service with observability
	inventoryConn, err := grpc.Dial(inventoryAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(observability.UnaryClientInterceptor(serviceName, logger)),
	)
	if err != nil {
		logger.Fatal("Failed to connect to inventory service", zap.String("address", inventoryAddr), zap.Error(err))
	}
	defer inventoryConn.Close()

	// Connect to payment service with observability
	paymentConn, err := grpc.Dial(paymentAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(observability.UnaryClientInterceptor(serviceName, logger)),
	)
	if err != nil {
		logger.Fatal("Failed to connect to payment service", zap.String("address", paymentAddr), zap.Error(err))
	}
	defer paymentConn.Close()

	// Create order service
	orderService := order.NewService(logger, inventoryConn, paymentConn)

	// Create gRPC server with observability interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(observability.UnaryServerInterceptor(serviceName, logger)),
		grpc.StreamInterceptor(observability.StreamServerInterceptor(serviceName, logger)),
	)

	orderServer := &orderServiceServer{
		service: orderService,
		logger:  logger,
	}

	// Register services
	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)
	reflection.Register(grpcServer)

	// Start metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", observability.MetricsHandler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		logger.Info("Starting metrics server", zap.String("port", metricsPort))
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.String("port", port), zap.Error(err))
	}

	logger.Info("Order Service listening",
		zap.String("grpc_address", lis.Addr().String()),
		zap.String("metrics_port", metricsPort))

	// Start server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Order Service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Order Service stopped gracefully")
	case <-ctx.Done():
		logger.Warn("Order Service shutdown timeout, forcing stop")
		grpcServer.Stop()
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
