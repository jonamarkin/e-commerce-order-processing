package domain_test // Use package_test for black-box testing

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
)

func TestNewOrder(t *testing.T) {
	customerID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()

	tests := []struct {
		name       string
		customerID uuid.UUID
		items      []domain.OrderItem
		wantErr    error
	}{
		{
			name:       "Successful order creation",
			customerID: customerID,
			items: []domain.OrderItem{
				{ProductID: productID1, Quantity: 1, UnitPrice: 10.0},
				{ProductID: productID2, Quantity: 2, UnitPrice: 5.0},
			},
			wantErr: nil,
		},
		{
			name:       "No order items",
			customerID: customerID,
			items:      []domain.OrderItem{},
			wantErr:    domain.ErrNoOrderItems,
		},
		{
			name:       "Order item with zero quantity",
			customerID: customerID,
			items: []domain.OrderItem{
				{ProductID: productID1, Quantity: 0, UnitPrice: 10.0},
			},
			wantErr: domain.ErrInvalidOrderItemQuantity,
		},
		{
			name:       "Order item with negative quantity",
			customerID: customerID,
			items: []domain.OrderItem{
				{ProductID: productID1, Quantity: -1, UnitPrice: 10.0},
			},
			wantErr: domain.ErrInvalidOrderItemQuantity,
		},
		{
			name:       "Order item with zero unit price",
			customerID: customerID,
			items: []domain.OrderItem{
				{ProductID: productID1, Quantity: 1, UnitPrice: 0.0},
			},
			wantErr: domain.ErrInvalidOrderItemUnitPrice,
		},
		{
			name:       "Order item with negative unit price",
			customerID: customerID,
			items: []domain.OrderItem{
				{ProductID: productID1, Quantity: 1, UnitPrice: -5.0},
			},
			wantErr: domain.ErrInvalidOrderItemUnitPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := domain.NewOrder(tt.customerID, tt.items)

			// Check for expected error
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("NewOrder() error = %v, wantErr %v", err, tt.wantErr)
				}
				if order != nil {
					t.Errorf("NewOrder() got order %v, want nil for error case", order)
				}
			} else {
				// Check for no error
				if err != nil {
					t.Errorf("NewOrder() unexpected error: %v", err)
				}
				// Check for successful order properties
				if order == nil {
					t.Error("NewOrder() returned nil order on success")
					return
				}
				if order.ID == uuid.Nil {
					t.Error("NewOrder() did not assign an ID")
				}
				if order.CustomerID != tt.customerID {
					t.Errorf("NewOrder() customer ID = %v, want %v", order.CustomerID, tt.customerID)
				}
				if domain.OrderStatus(order.Status) != domain.OrderStatusPending {
					t.Errorf("NewOrder() status = %v, want %v", order.Status, domain.OrderStatusPending)
				}
				if len(order.Items) != len(tt.items) {
					t.Errorf("NewOrder() item count = %d, want %d", len(order.Items), len(tt.items))
				}
				// Verify total price calculation
				expectedTotalPrice := 0.0
				for _, item := range tt.items {
					expectedTotalPrice += float64(item.Quantity) * item.UnitPrice
				}
				if order.TotalPrice != expectedTotalPrice {
					t.Errorf("NewOrder() total price = %f, want %f", order.TotalPrice, expectedTotalPrice)
				}
			}
		})
	}
}
