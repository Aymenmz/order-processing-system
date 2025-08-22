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

// 	"github.com/your-org/order-processing-system/pkg/order"
// 	orderpb "github.com/your-org/order-processing-system/pkg/pb/order"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// 	"google.golang.org/grpc/reflection"
// )

// // orderServiceServer implements the gRPC OrderService interface
// type orderServiceServer struct {
// 	orderpb.UnimplementedOrderServiceServer
// 	service order.Service
// 	logger  *zap.Logger
// }

// // CreateOrder handles order creation requests
// func (s *orderServiceServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
// 	s.logger.Info("Received CreateOrder request", zap.String("customer_id", req.CustomerId), zap.Int("items_count", len(req.Items)))

// 	order, err := s.service.CreateOrder(ctx, req.CustomerId, req.Items)
// 	if err != nil {
// 		s.logger.Error("Failed to create order", zap.Error(err))
// 		return nil, err
// 	}

// 	return &orderpb.CreateOrderResponse{Order: order}, nil
// }

// // GetOrder handles order retrieval requests
// func (s *orderServiceServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
// 	s.logger.Debug("Received GetOrder request", zap.String("order_id", req.OrderId))

// 	order, err := s.service.GetOrder(ctx, req.OrderId)
// 	if err != nil {
// 		s.logger.Error("Failed to get order", zap.String("order_id", req.OrderId), zap.Error(err))
// 		return nil, err
// 	}

// 	return &orderpb.GetOrderResponse{Order: order}, nil
// }

// // UpdateOrderStatus handles order status update requests
// func (s *orderServiceServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
// 	s.logger.Info("Received UpdateOrderStatus request", zap.String("order_id", req.OrderId), zap.String("new_status", req.NewStatus.String()))

// 	order, err := s.service.UpdateOrderStatus(ctx, req.OrderId, req.NewStatus)
// 	if err != nil {
// 		s.logger.Error("Failed to update order status", zap.String("order_id", req.OrderId), zap.Error(err))
// 		return nil, err
// 	}

// 	return &orderpb.UpdateOrderStatusResponse{Order: order}, nil
// }

// func main() {
// 	// Initialize logger
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	defer logger.Sync()

// 	logger.Info("Starting Order Service")

// 	// Get service addresses from environment variables
// 	inventoryAddr := getEnv("INVENTORY_SERVICE_ADDR", "localhost:50052")
// 	paymentAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50053")
// 	port := getEnv("PORT", "50051")

// 	// Connect to inventory service
// 	inventoryConn, err := grpc.Dial(inventoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		logger.Fatal("Failed to connect to inventory service", zap.String("address", inventoryAddr), zap.Error(err))
// 	}
// 	defer inventoryConn.Close()

// 	// Connect to payment service
// 	paymentConn, err := grpc.Dial(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		logger.Fatal("Failed to connect to payment service", zap.String("address", paymentAddr), zap.Error(err))
// 	}
// 	defer paymentConn.Close()

// 	// Create order service
// 	orderService := order.NewService(logger, inventoryConn, paymentConn)

// 	// Create gRPC server
// 	grpcServer := grpc.NewServer()
// 	orderServer := &orderServiceServer{
// 		service: orderService,
// 		logger:  logger,
// 	}

// 	// Register services
// 	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)
// 	reflection.Register(grpcServer)

// 	// Start listening
// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
// 	if err != nil {
// 		logger.Fatal("Failed to listen", zap.String("port", port), zap.Error(err))
// 	}

// 	logger.Info("Order Service listening", zap.String("address", lis.Addr().String()))

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

// 	logger.Info("Shutting down Order Service")

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
// 		logger.Info("Order Service stopped gracefully")
// 	case <-ctx.Done():
// 		logger.Warn("Order Service shutdown timeout, forcing stop")
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

