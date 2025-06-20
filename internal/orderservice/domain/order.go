package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID         uuid.UUID   `json:"id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	Status     string      `json:"status"`
	TotalPrice float64     `json:"total_price"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusFailed     OrderStatus = "failed"
)

func NewOrder(customerID uuid.UUID, items []OrderItem) (*Order, error) {
	//Calculate total price
	totalPrice := 0.0
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidOrderItemQuantity
		}
		if item.UnitPrice < 0 {
			return nil, ErrInvalidOrderItemUnitPrice
		}
		totalPrice += float64(item.Quantity) * item.UnitPrice

	}

	if len(items) == 0 {
		return nil, ErrNoOrderItems
	}

	now := time.Now()

	return &Order{
		ID:         uuid.New(),
		CustomerID: customerID,
		Items:      items,
		Status:     string(OrderStatusPending),
		TotalPrice: totalPrice,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
