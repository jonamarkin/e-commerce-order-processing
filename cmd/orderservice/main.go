package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/kafka"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/api"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/config"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/server"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/service"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env:", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: %+v\n", cfg)

	// --- Database Initialization ---
	db, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err = db.PingContext(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database!")

	// --- Kafka Producer Initialization ---
	const orderPlacedTopic = "orders.placed"
	kafkaProducer := kafka.NewProducer(cfg.KafkaBrokers, orderPlacedTopic)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Printf("Failed to close Kafka producer: %v", err)
		}
	}()
	log.Printf("Kafka producer initialized for topic: %s with brokers: %v", orderPlacedTopic, cfg.KafkaBrokers)

	// --- Initialize Repository, Service, and API Handler ---
	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, kafkaProducer)
	orderHandler := api.NewHandler(orderService)

	// --- Initialize and Run HTTP Server ---
	srv := server.NewServer(orderHandler, cfg.ServerPort)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}

func connectDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
