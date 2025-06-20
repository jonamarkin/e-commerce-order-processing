package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
)

type OrderRepository interface {
	// CreateOrder saves a new order to the repository.
	CreateOrder(ctx context.Context, order *domain.Order) error
	// GetOrderByID retrieves an order by its ID.
	GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	// UpdateOrderStatus updates the status of an existing order.
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
}
