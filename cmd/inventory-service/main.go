// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/your-org/order-processing-system/pkg/inventory"
// 	inventorypb "github.com/your-org/order-processing-system/pkg/pb/inventory"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/reflection"
// )

// // inventoryServiceServer implements the gRPC InventoryService interface
// type inventoryServiceServer struct {
// 	inventorypb.UnimplementedInventoryServiceServer
// 	service inventory.Service
// 	logger  *zap.Logger
// }

// // ReserveStock handles stock reservation requests
// func (s *inventoryServiceServer) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
// 	s.logger.Info("Received ReserveStock request", zap.String("product_id", req.ProductId), zap.Int32("quantity", req.Quantity), zap.String("order_id", req.OrderId))

// 	response, err := s.service.ReserveStock(ctx, req.ProductId, req.Quantity, req.OrderId)
// 	if err != nil {
// 		s.logger.Error("Failed to reserve stock", zap.Error(err))
// 		return nil, err
// 	}

// 	return response, nil
// }

// // ReleaseStock handles stock release requests
// func (s *inventoryServiceServer) ReleaseStock(ctx context.Context, req *inventorypb.ReleaseStockRequest) (*inventorypb.ReleaseStockResponse, error) {
// 	s.logger.Info("Received ReleaseStock request", zap.String("product_id", req.ProductId), zap.Int32("quantity", req.Quantity), zap.String("order_id", req.OrderId))

// 	response, err := s.service.ReleaseStock(ctx, req.ProductId, req.Quantity, req.OrderId)
// 	if err != nil {
// 		s.logger.Error("Failed to release stock", zap.Error(err))
// 		return nil, err
// 	}

// 	return response, nil
// }

// // GetProductStock handles product stock retrieval requests
// func (s *inventoryServiceServer) GetProductStock(ctx context.Context, req *inventorypb.GetProductStockRequest) (*inventorypb.GetProductStockResponse, error) {
// 	s.logger.Debug("Received GetProductStock request", zap.String("product_id", req.ProductId))

// 	product, err := s.service.GetProductStock(ctx, req.ProductId)
// 	if err != nil {
// 		s.logger.Error("Failed to get product stock", zap.String("product_id", req.ProductId), zap.Error(err))
// 		return nil, err
// 	}

// 	return &inventorypb.GetProductStockResponse{Product: product}, nil
// }

// func main() {
// 	// Initialize logger
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	defer logger.Sync()

// 	logger.Info("Starting Inventory Service")

// 	// Get port from environment variable
// 	port := getEnv("PORT", "50052")

// 	// Create inventory service
// 	inventoryService := inventory.NewService(logger)

// 	// Create gRPC server
// 	grpcServer := grpc.NewServer()
// 	inventoryServer := &inventoryServiceServer{
// 		service: inventoryService,
// 		logger:  logger,
// 	}

// 	// Register services
// 	inventorypb.RegisterInventoryServiceServer(grpcServer, inventoryServer)
// 	reflection.Register(grpcServer)

// 	// Start listening
// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
// 	if err != nil {
// 		logger.Fatal("Failed to listen", zap.String("port", port), zap.Error(err))
// 	}

// 	logger.Info("Inventory Service listening", zap.String("address", lis.Addr().String()))

// 	// Start server in a goroutine
// 	go func() {
// 		if err := grpcServer.Serve(lis); err != nil {
// 			logger.Fatal("Failed to serve gRPC server", zap.Error(err))
// 		}
// 	}()

// 	// Wait for interrupt signal to gracefully shutdown
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit

// 	logger.Info("Shutting down Inventory Service")

// 	// Graceful shutdown
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	done := make(chan struct{})
// 	go func() {
// 		grpcServer.GracefulStop()
// 		close(done)
// 	}()

// 	select {
// 	case <-done:
// 		logger.Info("Inventory Service stopped gracefully")
// 	case <-ctx.Done():
// 		logger.Warn("Inventory Service shutdown timeout, forcing stop")
// 		grpcServer.Stop()
// 	}
// }

// // getEnv gets an environment variable with a default value
// func getEnv(key, defaultValue string) string {
// 	if value := os.Getenv(key); value != "" {
// 		return value
// 	}
// 	return defaultValue
// }

