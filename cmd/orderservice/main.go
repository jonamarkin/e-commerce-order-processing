package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/config"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, proceeding with environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Configuration loaded successfully: %+v\n", cfg)

	// Database initialization
	db, err := connectDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	//Ping the database to ensure connection is established
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	log.Println("Database connection established successfully")

	//Initialize repository
	orderRepo := repository.NewPostgresOrderRepository(db)

	// Example usage of the repository
	ctx := context.Background()
	customerID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()

	newOrderItems := []domain.OrderItem{
		{ProductID: productID1, Quantity: 2, UnitPrice: 19.99},
		{ProductID: productID2, Quantity: 1, UnitPrice: 29.99},
	}

	newOrder, err := domain.NewOrder(customerID, newOrderItems)
	if err != nil {
		log.Fatalf("Error creating new order: %v", err)
	}

	log.Printf("Attempting to create a new order with ID: %s", newOrder.ID)

	err = orderRepo.CreateOrder(ctx, newOrder)
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)

	}
	log.Printf("Order created successfully with ID: %s", newOrder.ID)

	//get the order by ID
	fetchedOrder, err := orderRepo.GetOrderByID(ctx, newOrder.ID)
	if err != nil {
		log.Fatalf("Failed to fetch order by ID: %v", err)
	}
	log.Printf("Fetched order:` %+v", fetchedOrder)
	log.Printf("Fetched order items: %+v", fetchedOrder.Items)

	// Update order status
	log.Printf("Attempting to update order status for order ID: %s", newOrder.ID)
	err = orderRepo.UpdateOrderStatus(ctx, newOrder.ID, domain.OrderStatusProcessing)
	if err != nil {
		log.Fatalf("Failed to update order status: %v", err)
	}
	log.Printf("Order %s status updated to %s", newOrder.ID, domain.OrderStatusProcessing)

	//Verify the updated order status
	fetchedUpdatedOrder, err := orderRepo.GetOrderByID(ctx, newOrder.ID)
	if err != nil {
		log.Fatalf("Failed to fetch updated order by ID: %v", err)
	}
	log.Printf("Fetched updated order: %+v", fetchedUpdatedOrder)
	log.Printf("Fetched updated order status: %s", fetchedUpdatedOrder.Status)

	log.Println("Order service is running successfully")

}

func connectDatabase(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
