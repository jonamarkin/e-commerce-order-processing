package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/jonamarkin/e-commerce-order-processing/internal/inventoryservice/config"
	"github.com/jonamarkin/e-commerce-order-processing/internal/inventoryservice/kafka"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env:", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load Inventory Service configuration: %v", err)
	}

	log.Printf("Inventory Service Configuration loaded: %+v\n", cfg)

	// Initialize Kafka Consumer
	orderPlacedConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer func() {
		if err := orderPlacedConsumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled on main exit

	// Start consuming in a goroutine
	go orderPlacedConsumer.StartConsuming(ctx)

	// Listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-quit
	log.Println("Inventory Service: Shutting down...")
}
