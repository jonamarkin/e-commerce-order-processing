package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/kafka"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/api"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/config"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
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

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading configuration")
	}

	// --- Database Connection ---
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	// Ping database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}
	log.Info().Msg("Successfully connected to the database!")

	// --- Kafka Producer Initialization ---
	const orderPlacedTopic = "orders.placed"
	kafkaProducer := kafka.NewProducer(cfg.KafkaBrokers, orderPlacedTopic)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Kafka producer")
		}
	}()
	log.Info().Str("topic", orderPlacedTopic).Strs("brokers", cfg.KafkaBrokers).Msg("Kafka producer initialized")

	// --- Initialize Repository, Service, and API Handler ---
	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, kafkaProducer)
	orderHandler := api.NewHandler(orderService)

	// --- Gin Router Setup ---
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders/:id", orderHandler.GetOrderByID)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- HTTP Server Setup and Graceful Shutdown ---
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: router,
	}

	go func() {
		log.Info().Int("port", cfg.ServerPort).Msg("Server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to listen")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	log.Info().Msg("Server exited gracefully.")
}
