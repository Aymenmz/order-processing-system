package inventory

import (
	"context"
	"fmt"
	"sync"

	inventorypb "github.com/your-org/order-processing-system/pkg/pb/inventory"
	"go.uber.org/zap"
)

// Service defines the core inventory service interface
type Service interface {
	ReserveStock(ctx context.Context, productID string, quantity int32, orderID string) (*inventorypb.ReserveStockResponse, error)
	ReleaseStock(ctx context.Context, productID string, quantity int32, orderID string) (*inventorypb.ReleaseStockResponse, error)
	GetProductStock(ctx context.Context, productID string) (*inventorypb.Product, error)
}

// service implements the Service interface
type service struct {
	products map[string]*inventorypb.Product
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewService creates a new inventory service instance
func NewService(logger *zap.Logger) Service {
	// Initialize with some sample products
	products := map[string]*inventorypb.Product{
		"product-1": {
			Id:            "product-1",
			Name:          "Laptop",
			StockQuantity: 50,
			Price:         999.99,
		},
		"product-2": {
			Id:            "product-2",
			Name:          "Mouse",
			StockQuantity: 100,
			Price:         29.99,
		},
		"product-3": {
			Id:            "product-3",
			Name:          "Keyboard",
			StockQuantity: 75,
			Price:         79.99,
		},
	}

	return &service{
		products: products,
		logger:   logger,
	}
}

// ReserveStock reserves stock for a product
func (s *service) ReserveStock(ctx context.Context, productID string, quantity int32, orderID string) (*inventorypb.ReserveStockResponse, error) {
	s.logger.Info("Reserving stock", zap.String("product_id", productID), zap.Int32("quantity", quantity), zap.String("order_id", orderID))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	product, exists := s.products[productID]
	if !exists {
		s.logger.Warn("Product not found", zap.String("product_id", productID))
		return &inventorypb.ReserveStockResponse{
			Success: false,
			Message: fmt.Sprintf("Product not found: %s", productID),
		}, nil
	}

	if product.StockQuantity < quantity {
		s.logger.Warn("Insufficient stock", zap.String("product_id", productID), zap.Int32("available", product.StockQuantity), zap.Int32("requested", quantity))
		return &inventorypb.ReserveStockResponse{
			Success: false,
			Message: fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", product.StockQuantity, quantity),
		}, nil
	}

	// Reserve the stock
	product.StockQuantity -= quantity

	s.logger.Info("Stock reserved successfully", zap.String("product_id", productID), zap.Int32("reserved_quantity", quantity), zap.Int32("remaining_stock", product.StockQuantity))

	return &inventorypb.ReserveStockResponse{
		Success:          true,
		Message:          "Stock reserved successfully",
		ReservedQuantity: quantity,
	}, nil
}

// ReleaseStock releases previously reserved stock
func (s *service) ReleaseStock(ctx context.Context, productID string, quantity int32, orderID string) (*inventorypb.ReleaseStockResponse, error) {
	s.logger.Info("Releasing stock", zap.String("product_id", productID), zap.Int32("quantity", quantity), zap.String("order_id", orderID))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	product, exists := s.products[productID]
	if !exists {
		s.logger.Warn("Product not found for stock release", zap.String("product_id", productID))
		return &inventorypb.ReleaseStockResponse{
			Success: false,
			Message: fmt.Sprintf("Product not found: %s", productID),
		}, nil
	}

	// Release the stock
	product.StockQuantity += quantity

	s.logger.Info("Stock released successfully", zap.String("product_id", productID), zap.Int32("released_quantity", quantity), zap.Int32("current_stock", product.StockQuantity))

	return &inventorypb.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

// GetProductStock retrieves product information including stock
func (s *service) GetProductStock(ctx context.Context, productID string) (*inventorypb.Product, error) {
	s.logger.Debug("Getting product stock", zap.String("product_id", productID))

	s.mutex.RLock()
	product, exists := s.products[productID]
	s.mutex.RUnlock()

	if !exists {
		s.logger.Warn("Product not found", zap.String("product_id", productID))
		return nil, fmt.Errorf("product not found: %s", productID)
	}

	return product, nil
}

