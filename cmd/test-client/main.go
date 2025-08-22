package main

import (
	"context"
	"log"
	"time"

	orderpb "github.com/your-org/order-processing-system/pkg/pb/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to order service
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to order service: %v", err)
	}
	defer conn.Close()

	client := orderpb.NewOrderServiceClient(conn)

	// Create a test order
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orderReq := &orderpb.CreateOrderRequest{
		CustomerId: "customer-123",
		Items: []*orderpb.OrderItem{
			{
				ProductId: "product-1",
				Quantity:  2,
				UnitPrice: 999.99,
			},
			{
				ProductId: "product-2",
				Quantity:  1,
				UnitPrice: 29.99,
			},
		},
	}

	log.Println("Creating order...")
	orderResp, err := client.CreateOrder(ctx, orderReq)
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)
	}

	log.Printf("Order created successfully: %s", orderResp.Order.Id)
	log.Printf("Total amount: $%.2f", orderResp.Order.TotalAmount)
	log.Printf("Status: %s", orderResp.Order.Status.String())

	// Get the order
	log.Println("Retrieving order...")
	getReq := &orderpb.GetOrderRequest{OrderId: orderResp.Order.Id}
	getResp, err := client.GetOrder(ctx, getReq)
	if err != nil {
		log.Fatalf("Failed to get order: %v", err)
	}

	log.Printf("Retrieved order: %s", getResp.Order.Id)
	log.Printf("Customer: %s", getResp.Order.CustomerId)
	log.Printf("Items count: %d", len(getResp.Order.Items))

	// Update order status
	log.Println("Updating order status...")
	updateReq := &orderpb.UpdateOrderStatusRequest{
		OrderId:   orderResp.Order.Id,
		NewStatus: orderpb.OrderStatus_ORDER_STATUS_COMPLETED,
	}
	updateResp, err := client.UpdateOrderStatus(ctx, updateReq)
	if err != nil {
		log.Fatalf("Failed to update order status: %v", err)
	}

	log.Printf("Order status updated to: %s", updateResp.Order.Status.String())
	log.Println("Test completed successfully!")
}

