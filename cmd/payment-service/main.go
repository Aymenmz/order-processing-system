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

// 	"github.com/your-org/order-processing-system/pkg/payment"
// 	paymentpb "github.com/your-org/order-processing-system/pkg/pb/payment"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/reflection"
// )

// // paymentServiceServer implements the gRPC PaymentService interface
// type paymentServiceServer struct {
// 	paymentpb.UnimplementedPaymentServiceServer
// 	service payment.Service
// 	logger  *zap.Logger
// }

// // ProcessPayment handles payment processing requests
// func (s *paymentServiceServer) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
// 	s.logger.Info("Received ProcessPayment request",
// 		zap.String("order_id", req.OrderId),
// 		zap.String("customer_id", req.CustomerId),
// 		zap.Float64("amount", req.Amount))

// 	response, err := s.service.ProcessPayment(ctx, req)
// 	if err != nil {
// 		s.logger.Error("Failed to process payment", zap.Error(err))
// 		return nil, err
// 	}

// 	return response, nil
// }

// func main() {
// 	// Initialize logger
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	defer logger.Sync()

// 	logger.Info("Starting Payment Service")

// 	// Get port from environment variable
// 	port := getEnv("PORT", "50053")

// 	// Create payment service
// 	paymentService := payment.NewService(logger)

// 	// Create gRPC server
// 	grpcServer := grpc.NewServer()
// 	paymentServer := &paymentServiceServer{
// 		service: paymentService,
// 		logger:  logger,
// 	}

// 	// Register services
// 	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentServer)
// 	reflection.Register(grpcServer)

// 	// Start listening
// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
// 	if err != nil {
// 		logger.Fatal("Failed to listen", zap.String("port", port), zap.Error(err))
// 	}

// 	logger.Info("Payment Service listening", zap.String("address", lis.Addr().String()))

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

// 	logger.Info("Shutting down Payment Service")

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
// 		logger.Info("Payment Service stopped gracefully")
// 	case <-ctx.Done():
// 		logger.Warn("Payment Service shutdown timeout, forcing stop")
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

