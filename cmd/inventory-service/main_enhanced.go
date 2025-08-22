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

	"github.com/your-org/order-processing-system/pkg/inventory"
	"github.com/your-org/order-processing-system/pkg/observability"
	inventorypb "github.com/your-org/order-processing-system/pkg/pb/inventory"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// inventoryServiceServer implements the gRPC InventoryService interface with observability
type inventoryServiceServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	service inventory.Service
	logger  *zap.Logger
}

// ReserveStock handles stock reservation requests with enhanced logging and metrics
func (s *inventoryServiceServer) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	contextLogger := observability.LoggerWithOrderID(
		observability.LoggerWithTraceContext(ctx, s.logger),
		req.OrderId,
	)

	contextLogger.Info("Processing ReserveStock request",
		zap.String("product_id", req.ProductId),
		zap.Int32("quantity", req.Quantity))

	response, err := s.service.ReserveStock(ctx, req.ProductId, req.Quantity, req.OrderId)
	if err != nil {
		contextLogger.Error("Failed to reserve stock", zap.Error(err))
		observability.InventoryReservations.WithLabelValues(req.ProductId, "error").Inc()
		return nil, err
	}

	// Record metrics based on success/failure
	status := "success"
	if !response.Success {
		status = "failed"
	}
	observability.InventoryReservations.WithLabelValues(req.ProductId, status).Inc()

	contextLogger.Info("Stock reservation processed",
		zap.Bool("success", response.Success),
		zap.String("message", response.Message))

	return response, nil
}

// ReleaseStock handles stock release requests
func (s *inventoryServiceServer) ReleaseStock(ctx context.Context, req *inventorypb.ReleaseStockRequest) (*inventorypb.ReleaseStockResponse, error) {
	contextLogger := observability.LoggerWithOrderID(
		observability.LoggerWithTraceContext(ctx, s.logger),
		req.OrderId,
	)

	contextLogger.Info("Processing ReleaseStock request",
		zap.String("product_id", req.ProductId),
		zap.Int32("quantity", req.Quantity))

	response, err := s.service.ReleaseStock(ctx, req.ProductId, req.Quantity, req.OrderId)
	if err != nil {
		contextLogger.Error("Failed to release stock", zap.Error(err))
		return nil, err
	}

	contextLogger.Info("Stock release processed",
		zap.Bool("success", response.Success),
		zap.String("message", response.Message))

	return response, nil
}

// GetProductStock handles product stock retrieval requests
func (s *inventoryServiceServer) GetProductStock(ctx context.Context, req *inventorypb.GetProductStockRequest) (*inventorypb.GetProductStockResponse, error) {
	contextLogger := observability.LoggerWithTraceContext(ctx, s.logger)
	contextLogger.Debug("Processing GetProductStock request", zap.String("product_id", req.ProductId))

	product, err := s.service.GetProductStock(ctx, req.ProductId)
	if err != nil {
		contextLogger.Error("Failed to get product stock", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, err
	}

	// Update current stock gauge metric
	observability.CurrentStock.WithLabelValues(product.Id, product.Name).Set(float64(product.StockQuantity))

	contextLogger.Debug("Product stock retrieved",
		zap.String("product_name", product.Name),
		zap.Int32("stock_quantity", product.StockQuantity))

	return &inventorypb.GetProductStockResponse{Product: product}, nil
}

func main() {
	serviceName := "inventory-service"

	// Initialize structured logger
	logger, err := observability.NewLogger(serviceName, zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Inventory Service with observability")

	// Initialize metrics
	observability.InitMetrics()

	// Initialize tracing
	jaegerEndpoint := getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	cleanup, err := observability.InitTracing(serviceName, jaegerEndpoint, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracing", zap.Error(err))
	}
	defer cleanup()

	// Get configuration from environment variables
	port := getEnv("PORT", "50052")
	metricsPort := getEnv("METRICS_PORT", "8081")

	// Create inventory service
	inventoryService := inventory.NewService(logger)

	// Create gRPC server with observability interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(observability.UnaryServerInterceptor(serviceName, logger)),
		grpc.StreamInterceptor(observability.StreamServerInterceptor(serviceName, logger)),
	)

	inventoryServer := &inventoryServiceServer{
		service: inventoryService,
		logger:  logger,
	}

	// Register services
	inventorypb.RegisterInventoryServiceServer(grpcServer, inventoryServer)
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

	logger.Info("Inventory Service listening",
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

	logger.Info("Shutting down Inventory Service")

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
		logger.Info("Inventory Service stopped gracefully")
	case <-ctx.Done():
		logger.Warn("Inventory Service shutdown timeout, forcing stop")
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
