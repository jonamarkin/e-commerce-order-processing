package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/api"
)

// Server represents the HTTP server for the Order Service.
type Server struct {
	router *gin.Engine
	port   int
}

// NewServer creates a new Server instance
func NewServer(handler *api.Handler, port int) *Server {
	router := gin.Default()

	// Middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next() // Process request
		duration := time.Since(start)
		log.Printf("Request - Method: %s, Path: %s, Status: %d, Duration: %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	})

	// --- Health Check Endpoint ---
	router.GET("/health", handler.HealthCheck)

	// --- API Grouping ---
	apiV1 := router.Group("/api/v1")
	{
		ordersGroup := apiV1.Group("/orders")
		{
			ordersGroup.POST("/", handler.CreateOrder)
			ordersGroup.GET("/:id", handler.GetOrderByID)
		}
	}

	return &Server{
		router: router,
		port:   port,
	}
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create a channel to listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	// Notify quit channel on SIGINT (Ctrl+C) and SIGTERM (Docker/Kubernetes shutdown)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine so it doesn't block
	go func() {
		log.Printf("Order Service starting on port %d...", s.port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to listen: %v", err)
		}
	}()

	// Block until a signal is received
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
	return nil
}
