package service_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, orderID)
	// The first return value (0) is the *domain.Order, second (1) is the error
	return args.Get(0).(*domain.Order), args.Error(1)
}

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) PublishMessage(ctx context.Context, key, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}
