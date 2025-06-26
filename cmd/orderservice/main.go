package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
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

	_ "github.com/jonamarkin/e-commerce-order-processing/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title E-Commerce Order Processing Service API
// @version 1.0
// @description This is the API documentation for the E-Commerce Order Processing Service.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

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

	// --- Gin Router Setup ---
	router := gin.Default()

	// Add a base group for versioning (recommended practice)
	v1 := router.Group("/api/v1")
	{
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders/:id", orderHandler.GetOrderByID)
	}

	// Swagger UI endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- Initialize and Run HTTP Server ---
	srv := server.NewServer(router, cfg.ServerPort)
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
