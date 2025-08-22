package order

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	orderpb "github.com/your-org/order-processing-system/pkg/pb/order"
	inventrypb "github.com/your-org/order-processing-system/pkg/pb/inventory"
	paymentpb "github.com/your-org/order-processing-system/pkg/pb/payment"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Service defines the core order service interface
type Service interface {
	CreateOrder(ctx context.Context, customerID string, items []*orderpb.OrderItem) (*orderpb.Order, error)
	GetOrder(ctx context.Context, orderID string) (*orderpb.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status orderpb.OrderStatus) (*orderpb.Order, error)
}

// service implements the Service interface
type service struct {
	orders           map[string]*orderpb.Order
	mutex            sync.RWMutex
	logger           *zap.Logger
	inventoryClient  inventrypb.InventoryServiceClient
	paymentClient    paymentpb.PaymentServiceClient
}

// NewService creates a new order service instance
func NewService(logger *zap.Logger, inventoryConn, paymentConn *grpc.ClientConn) Service {
	return &service{
		orders:          make(map[string]*orderpb.Order),
		logger:          logger,
		inventoryClient: inventrypb.NewInventoryServiceClient(inventoryConn),
		paymentClient:   paymentpb.NewPaymentServiceClient(paymentConn),
	}
}

// CreateOrder creates a new order
func (s *service) CreateOrder(ctx context.Context, customerID string, items []*orderpb.OrderItem) (*orderpb.Order, error) {
	s.logger.Info("Creating new order", zap.String("customer_id", customerID), zap.Int("items_count", len(items)))

	// Generate order ID
	orderID := uuid.New().String()

	// Calculate total amount
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.UnitPrice * float64(item.Quantity)
	}

	// Create order object
	order := &orderpb.Order{
		Id:          orderID,
		CustomerId:  customerID,
		Items:       items,
		TotalAmount: totalAmount,
		Status:      orderpb.OrderStatus_ORDER_STATUS_PENDING,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Reserve inventory for each item
	for _, item := range items {
		reserveReq := &inventrypb.ReserveStockRequest{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
			OrderId:   orderID,
		}

		reserveResp, err := s.inventoryClient.ReserveStock(ctx, reserveReq)
		if err != nil {
			s.logger.Error("Failed to reserve stock", zap.String("order_id", orderID), zap.String("product_id", item.ProductId), zap.Error(err))
			return nil, fmt.Errorf("failed to reserve stock for product %s: %w", item.ProductId, err)
		}

		if !reserveResp.Success {
			s.logger.Warn("Stock reservation failed", zap.String("order_id", orderID), zap.String("product_id", item.ProductId), zap.String("message", reserveResp.Message))
			return nil, fmt.Errorf("insufficient stock for product %s: %s", item.ProductId, reserveResp.Message)
		}
	}

	// Process payment
	paymentReq := &paymentpb.PaymentRequest{
		OrderId:       orderID,
		CustomerId:    customerID,
		Amount:        totalAmount,
		Currency:      "USD",
		PaymentMethod: "credit_card",
	}

	paymentResp, err := s.paymentClient.ProcessPayment(ctx, paymentReq)
	if err != nil {
		s.logger.Error("Payment processing failed", zap.String("order_id", orderID), zap.Error(err))
		// Release reserved stock
		s.releaseStockForOrder(ctx, orderID, items)
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	if paymentResp.Status != paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS {
		s.logger.Warn("Payment failed", zap.String("order_id", orderID), zap.String("message", paymentResp.Message))
		// Release reserved stock
		s.releaseStockForOrder(ctx, orderID, items)
		return nil, fmt.Errorf("payment failed: %s", paymentResp.Message)
	}

	// Update order status to processing
	order.Status = orderpb.OrderStatus_ORDER_STATUS_PROCESSING
	order.UpdatedAt = time.Now().Format(time.RFC3339)

	// Store order
	s.mutex.Lock()
	s.orders[orderID] = order
	s.mutex.Unlock()

	s.logger.Info("Order created successfully", zap.String("order_id", orderID), zap.Float64("total_amount", totalAmount))
	return order, nil
}

// GetOrder retrieves an order by ID
func (s *service) GetOrder(ctx context.Context, orderID string) (*orderpb.Order, error) {
	s.logger.Debug("Retrieving order", zap.String("order_id", orderID))

	s.mutex.RLock()
	order, exists := s.orders[orderID]
	s.mutex.RUnlock()

	if !exists {
		s.logger.Warn("Order not found", zap.String("order_id", orderID))
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	return order, nil
}

// UpdateOrderStatus updates the status of an order
func (s *service) UpdateOrderStatus(ctx context.Context, orderID string, status orderpb.OrderStatus) (*orderpb.Order, error) {
	s.logger.Info("Updating order status", zap.String("order_id", orderID), zap.String("new_status", status.String()))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		s.logger.Warn("Order not found for status update", zap.String("order_id", orderID))
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	order.Status = status
	order.UpdatedAt = time.Now().Format(time.RFC3339)

	s.logger.Info("Order status updated", zap.String("order_id", orderID), zap.String("status", status.String()))
	return order, nil
}

// releaseStockForOrder releases reserved stock for an order (helper function)
func (s *service) releaseStockForOrder(ctx context.Context, orderID string, items []*orderpb.OrderItem) {
	for _, item := range items {
		releaseReq := &inventrypb.ReleaseStockRequest{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
			OrderId:   orderID,
		}

		_, err := s.inventoryClient.ReleaseStock(ctx, releaseReq)
		if err != nil {
			s.logger.Error("Failed to release stock", zap.String("order_id", orderID), zap.String("product_id", item.ProductId), zap.Error(err))
		}
	}
}

