package payment

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	paymentpb "github.com/your-org/order-processing-system/pkg/pb/payment"
	"go.uber.org/zap"
)

// Service defines the core payment service interface
type Service interface {
	ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error)
}

// service implements the Service interface
type service struct {
	logger *zap.Logger
	rand   *rand.Rand
}

// NewService creates a new payment service instance
func NewService(logger *zap.Logger) Service {
	return &service{
		logger: logger,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ProcessPayment processes a payment request
func (s *service) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	s.logger.Info("Processing payment", 
		zap.String("order_id", req.OrderId), 
		zap.String("customer_id", req.CustomerId), 
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
		zap.String("payment_method", req.PaymentMethod))

	// Generate payment ID and transaction ID
	paymentID := uuid.New().String()
	transactionID := fmt.Sprintf("txn_%d", time.Now().Unix())

	// Simulate payment processing delay
	time.Sleep(time.Millisecond * time.Duration(s.rand.Intn(500)+100))

	// Simulate payment success/failure (90% success rate)
	success := s.rand.Float32() < 0.9

	if success {
		s.logger.Info("Payment processed successfully", 
			zap.String("payment_id", paymentID), 
			zap.String("transaction_id", transactionID),
			zap.String("order_id", req.OrderId))

		return &paymentpb.PaymentResponse{
			PaymentId:     paymentID,
			Status:        paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS,
			Message:       "Payment processed successfully",
			TransactionId: transactionID,
		}, nil
	} else {
		s.logger.Warn("Payment processing failed", 
			zap.String("payment_id", paymentID), 
			zap.String("order_id", req.OrderId))

		return &paymentpb.PaymentResponse{
			PaymentId:     paymentID,
			Status:        paymentpb.PaymentStatus_PAYMENT_STATUS_FAILED,
			Message:       "Payment declined by bank",
			TransactionId: transactionID,
		}, nil
	}
}

