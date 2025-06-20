package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/jonamarkin/e-commerce-order-processing/orderservice/internal/orderservice/config"
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
	fmt.Printf("Server will run on port: %d\n", cfg.ServerPort)
	fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("Kafka Brokers: %v\n", cfg.KafkaBrokers)

	log.Println("Order Service is starting...")

}
