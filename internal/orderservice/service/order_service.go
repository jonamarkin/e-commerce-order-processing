package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, customerID uuid.UUID, items []domain.OrderItem) (*domain.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
}

type orderServiceImpl struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderServiceImpl{
		orderRepo: repo,
	}
}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, customerID uuid.UUID, items []domain.OrderItem) (*domain.Order, error) {
	order, err := domain.NewOrder(customerID, items)
	if err != nil {
		log.Printf("Error creating new order: %v", err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		log.Printf("Error saving order to repository: %v", err)
		return nil, fmt.Errorf("failed to save order: %w", err)
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
