package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/kafka"
	"time"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
	"github.com/rs/zerolog/log"
)

type OrderService interface {
	CreateOrder(ctx context.Context, customerID uuid.UUID, items []domain.OrderItem) (*domain.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
}

type orderServiceImpl struct {
	orderRepo     repository.OrderRepository
	kafkaProducer kafka.KafkaProducer
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(repo repository.OrderRepository, producer kafka.KafkaProducer) OrderService {
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
		log.Ctx(ctx).Error().Err(err).Msg("Service: failed to create new order domain object") // Contextual logging
		return nil, fmt.Errorf("service: failed to create new order domain object: %w", err)
	}

	err = s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Service: failed to persist order")
		return nil, fmt.Errorf("service: failed to persist order: %w", err)
	}

	orderPlacedEvent := struct {
		OrderID    uuid.UUID `json:"order_id"`
		CustomerID uuid.UUID `json:"customer_id"`
		TotalPrice float64   `json:"total_price"`
		Timestamp  time.Time `json:"timestamp"`
		Items      []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
			UnitPrice float64   `json:"unit_price"`
		} `json:"items"`
	}{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		TotalPrice: order.TotalPrice,
		Timestamp:  order.CreatedAt,
	}

	for _, item := range order.Items {
		orderPlacedEvent.Items = append(orderPlacedEvent.Items, struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
			UnitPrice float64   `json:"unit_price"`
		}{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	eventValue, err := json.Marshal(orderPlacedEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Str("order_id", order.ID.String()).
			Msg("Service: Failed to marshal order placed event")
		return order, nil
	}

	err = s.kafkaProducer.PublishMessage(ctx, []byte(order.ID.String()), eventValue)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Str("order_id", order.ID.String()).
			Msg("Service: Failed to publish order placed event to Kafka")
		return order, nil
	}

	log.Ctx(ctx).Info().
		Str("order_id", order.ID.String()).
		Msg("Order created and 'orders.placed' event published to Kafka.")
	return order, nil
}

func (s *orderServiceImpl) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Str("order_id", orderID.String()).
			Msg("Service: failed to get order by ID")
		return nil, fmt.Errorf("service: failed to get order by ID %s: %w", orderID, err)
	}
	log.Ctx(ctx).Info().Str("order_id", orderID.String()).Msg("Order retrieved successfully")
	return order, nil
}
