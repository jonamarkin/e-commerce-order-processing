package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/kafka"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, customerID uuid.UUID, items []domain.OrderItem) (*domain.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
}

type orderServiceImpl struct {
	orderRepo     repository.OrderRepository
	kafkaProducer *kafka.Producer
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(repo repository.OrderRepository, producer *kafka.Producer) OrderService {
	return &orderServiceImpl{
		orderRepo:     repo,
		kafkaProducer: producer,
	}
}

type OrderPlacedEvent struct {
	OrderID    uuid.UUID          `json:"order_id"`
	CustomerID uuid.UUID          `json:"customer_id"`
	TotalPrice float64            `json:"total_price"`
	Status     domain.OrderStatus `json:"status"`
}

// CreateOrder handles the creation of a new order, applying business rules,
// persisting it, and publishing an event.
func (s *orderServiceImpl) CreateOrder(ctx context.Context, customerID uuid.UUID, items []domain.OrderItem) (*domain.Order, error) {
	order, err := domain.NewOrder(customerID, items)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create new order domain object: %w", err)
	}

	err = s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("service: failed to persist order: %w", err)
	}

	event := OrderPlacedEvent{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		TotalPrice: order.TotalPrice,
		Status:     domain.OrderStatus(order.Status),
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		// Log the error but don't fail the order creation, as the order is already persisted.
		// This is a trade-off: guaranteed delivery vs. eventual consistency.
		// For strong guarantees, consider the Outbox Pattern.
		log.Printf("Service: Failed to marshal OrderPlacedEvent for order %s: %v", order.ID, err)
	} else {
		// Use a separate context for Kafka publishing to allow it to complete even if HTTP request context times out
		kafkaCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Use order.ID as the Kafka message key to ensure messages for the same order go to the same partition
		err = s.kafkaProducer.PublishMessage(kafkaCtx, []byte(order.ID.String()), eventBytes)
		if err != nil {
			log.Printf("Service: Failed to publish OrderPlaced event for order %s: %v", order.ID, err)
		} else {
			log.Printf("Service: Successfully published OrderPlaced event for Order ID: %s", order.ID)
		}
	}

	return order, nil
}

func (s *orderServiceImpl) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}
