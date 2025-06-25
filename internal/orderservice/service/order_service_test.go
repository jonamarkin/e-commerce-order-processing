package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()
	customerID := uuid.New()
	productID := uuid.New()
	items := []domain.OrderItem{
		{ProductID: productID, Quantity: 2, UnitPrice: 10.0},
	}

	t.Run("successful order creation and event publishing", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockProducer := new(MockKafkaProducer)
		orderService := service.NewOrderService(mockRepo, mockProducer)

		mockRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil).Once()
		mockProducer.On("PublishMessage", mock.Anything, mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8")).Return(nil).Once()

		order, err := orderService.CreateOrder(ctx, customerID, items)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, customerID, order.CustomerID)
		assert.Equal(t, domain.OrderStatusPending, order.Status) // This should now pass due to domain.Order struct change
		assert.Len(t, order.Items, 1)
		assert.Equal(t, 20.0, order.TotalPrice)

		mockRepo.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failed to create order in repository", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)                            // NEW MOCK
		mockProducer := new(MockKafkaProducer)                          // NEW MOCK
		orderService := service.NewOrderService(mockRepo, mockProducer) // NEW SERVICE

		mockRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(errors.New("db error")).Once()

		order, err := orderService.CreateOrder(ctx, customerID, items)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "failed to persist order")

		mockRepo.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishMessage", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("failed to publish event (order still created)", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)                            // NEW MOCK
		mockProducer := new(MockKafkaProducer)                          // NEW MOCK
		orderService := service.NewOrderService(mockRepo, mockProducer) // NEW SERVICE

		mockRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil).Once()
		mockProducer.On("PublishMessage", mock.Anything, mock.AnythingOfType("[]uint8"), mock.AnythingOfType("[]uint8")).Return(errors.New("kafka error")).Once()

		order, err := orderService.CreateOrder(ctx, customerID, items)

		assert.NoError(t, err)
		assert.NotNil(t, order)

		mockRepo.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})
}

func TestOrderService_GetOrderByID(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()
	expectedOrder := &domain.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     domain.OrderStatusPending, // Using domain.OrderStatus here
		TotalPrice: 100.0,
		Items: []domain.OrderItem{
			{ProductID: productID, Quantity: 1, UnitPrice: 100.0},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)                            // NEW MOCK
		mockProducer := new(MockKafkaProducer)                          // NEW MOCK
		orderService := service.NewOrderService(mockRepo, mockProducer) // NEW SERVICE

		mockRepo.On("GetOrderByID", mock.Anything, orderID).Return(expectedOrder, nil).Once()

		order, err := orderService.GetOrderByID(ctx, orderID)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, expectedOrder, order)

		mockRepo.AssertExpectations(t)
	})

	t.Run("order not found", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)                            // NEW MOCK
		mockProducer := new(MockKafkaProducer)                          // NEW MOCK
		orderService := service.NewOrderService(mockRepo, mockProducer) // NEW SERVICE

		mockRepo.On("GetOrderByID", mock.Anything, orderID).Return(&domain.Order{}, domain.ErrOrderNotFound).Once()

		order, err := orderService.GetOrderByID(ctx, orderID)

		assert.ErrorIs(t, err, domain.ErrOrderNotFound)
		assert.Nil(t, order)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository returns generic error", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)                            // NEW MOCK
		mockProducer := new(MockKafkaProducer)                          // NEW MOCK
		orderService := service.NewOrderService(mockRepo, mockProducer) // NEW SERVICE

		repoError := errors.New("database connection lost")
		mockRepo.On("GetOrderByID", mock.Anything, orderID).Return(&domain.Order{}, repoError).Once()

		order, err := orderService.GetOrderByID(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "service: failed to get order by ID")
		assert.ErrorContains(t, err, repoError.Error())

		mockRepo.AssertExpectations(t)
	})
}
