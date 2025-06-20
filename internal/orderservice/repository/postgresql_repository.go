package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	_ "github.com/lib/pq"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new instance of PostgresOrderRepository.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// CreateOrder saves a new order and its items to the PostgreSQL database.
func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil) // Start a transaction for atomicity
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the order
	orderSQL := `
		INSERT INTO orders (id, customer_id, status, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx, orderSQL, order.ID, order.CustomerID, order.Status, order.TotalPrice, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert each order item
	orderItemSQL := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for _, item := range order.Items {
		itemID := uuid.New() // Generate a new UUID for the order item
		_, err = tx.ExecContext(ctx, orderItemSQL, itemID, order.ID, item.ProductID, item.Quantity, item.UnitPrice, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return tx.Commit() // Commit the transaction
}

// GetOrderByID retrieves an order by its ID from the PostgreSQL database.
func (r *PostgresOrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order := &domain.Order{}
	orderSQL := `
		SELECT id, customer_id, status, total_price, created_at, updated_at
		FROM orders
		WHERE id = $1`
	err := r.db.QueryRowContext(ctx, orderSQL, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Status,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	//Fetch order items
	itemSQL := `
		SELECT product_id, quantity, unit_price
		FROM order_items
		WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, itemSQL, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over order items: %w", err)
	}
	order.Items = items
	return order, nil

}

// UpdateOrderStatus updates the status of an existing order in the PostgreSQL database.
func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}
