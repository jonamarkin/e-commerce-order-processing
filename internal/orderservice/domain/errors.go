package domain

import "errors"

var (
	ErrInvalidOrderItemQuantity     = errors.New("invalid order item quantity")
	ErrInvalidOrderItemUnitPrice    = errors.New("invalid order item unit price")
	ErrNoOrderItems                 = errors.New("no order items provided")
	ErrOrderNotFound                = errors.New("order not found")
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")
)
