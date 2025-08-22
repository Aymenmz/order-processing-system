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
	"github.com/your-org/order-processing-system/pkg/payment"
	paymentpb "github.com/your-org/order-processing-system/pkg/pb/payment"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// paymentServiceServer implements the gRPC PaymentService interface with observability
type paymentServiceServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	service payment.Service
	logger  *zap.Logger
}

// ProcessPayment handles payment processing requests with enhanced logging and metrics
func (s *paymentServiceServer) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	contextLogger := observability.LoggerWithCustomerID(
		observability.LoggerWithOrderID(
			observability.LoggerWithTraceContext(ctx, s.logger),
			req.OrderId,
		),
		req.CustomerId,
	)

	contextLogger.Info("Processing payment request",
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
		zap.String("payment_method", req.PaymentMethod))

	response, err := s.service.ProcessPayment(ctx, req)
	if err != nil {
		contextLogger.Error("Failed to process payment", zap.Error(err))
		observability.PaymentsProcessed.WithLabelValues("error").Inc()
		return nil, err
	}

	// Record metrics based on payment status
	status := "success"
	if response.Status != paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS {
		status = "failed"
	}
	observability.PaymentsProcessed.WithLabelValues(status).Inc()

	contextLogger.Info("Payment processed",
		zap.String("payment_id", response.PaymentId),
		zap.String("status", response.Status.String()),
		zap.String("transaction_id", response.TransactionId))

	return response, nil
}

func main() {
	serviceName := "payment-service"

	// Initialize structured logger
	logger, err := observability.NewLogger(serviceName, zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Payment Service with observability")

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
	port := getEnv("PORT", "50053")
	metricsPort := getEnv("METRICS_PORT", "8082")

	// Create payment service
	paymentService := payment.NewService(logger)

	// Create gRPC server with observability interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(observability.UnaryServerInterceptor(serviceName, logger)),
		grpc.StreamInterceptor(observability.StreamServerInterceptor(serviceName, logger)),
	)

	paymentServer := &paymentServiceServer{
		service: paymentService,
		logger:  logger,
	}

	// Register services
	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentServer)
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

	logger.Info("Payment Service listening",
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

	logger.Info("Shutting down Payment Service")

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
		logger.Info("Payment Service stopped gracefully")
	case <-ctx.Done():
		logger.Warn("Payment Service shutdown timeout, forcing stop")
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
