package repository_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/repository"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// TestMain runs before all tests in this package.
func TestMain(m *testing.M) {
	// Load environment variables for DATABASE_URL (from .env or Docker Compose if running inside)
	// For integration tests, we need to ensure the database is accessible.
	// If running this directly via `go test`, DATABASE_URL from .env should point to localhost.
	// If running inside a test container, it would point to the 'db' service.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set for integration tests.")
	}

	var err error
	testDB, err = connectTestDB(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	// Run migrations (if not already done) - crucial for integration tests
	// This ensures the DB schema is ready for tests.
	// In a real setup, you might use a migration library directly here.
	// For simplicity, we assume migrations are applied before running tests
	// via `go run cmd/orderservice/migrate/main.go up` outside of this test run.
	// If you want to automate migrations *within* the test, it's more complex
	// (e.g., using testcontainers-go or a separate binary call).

	// Run all tests
	exitCode := m.Run()

	// Teardown: Clean up database after all tests
	if err := cleanupDatabase(testDB); err != nil {
		log.Printf("Failed to clean up test database: %v", err)
	}

	os.Exit(exitCode)
}

// connectTestDB establishes a connection to the test database.
func connectTestDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5) // Keep test connections low
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// clearTable clears the test tables before each test case (important for isolated tests).
func clearTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM order_items; DELETE FROM orders;")
	return err
}

// cleanupDatabase drops tables after all tests in TestMain.
func cleanupDatabase(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS order_items; DROP TABLE IF EXISTS orders;")
	return err
}

func TestPostgresOrderRepository(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available for integration tests")
	}

	repo := repository.NewPostgresOrderRepository(testDB)
	ctx := context.Background()

	// Run tests in parallel for isolation
	t.Parallel()

	// Before each test in this suite, clear tables
	t.Cleanup(func() {
		if err := clearTable(testDB); err != nil {
			t.Fatalf("Failed to clear table: %v", err)
		}
	})

	t.Run("Create and Get Order", func(t *testing.T) {
		t.Parallel() // Allow this subtest to run in parallel with other parallel subtests

		customerID := uuid.New()
		productID1 := uuid.New()
		productID2 := uuid.New()

		items := []domain.OrderItem{
			{ProductID: productID1, Quantity: 1, UnitPrice: 10.0},
			{ProductID: productID2, Quantity: 2, UnitPrice: 5.0},
		}

		order, err := domain.NewOrder(customerID, items)
		assert.NoError(t, err)
		assert.NotNil(t, order)

		// Create the order
		err = repo.CreateOrder(ctx, order)
		assert.NoError(t, err, "Expected no error when creating order")

		// Get the order back
		retrievedOrder, err := repo.GetOrderByID(ctx, order.ID)
		assert.NoError(t, err, "Expected no error when getting order by ID")
		assert.NotNil(t, retrievedOrder, "Expected a retrieved order, got nil")

		// Assert properties of the retrieved order
		assert.Equal(t, order.ID, retrievedOrder.ID)
		assert.Equal(t, order.CustomerID, retrievedOrder.CustomerID)
		assert.Equal(t, order.Status, retrievedOrder.Status)
		assert.InDelta(t, order.TotalPrice, retrievedOrder.TotalPrice, 0.001) // Use InDelta for floats
		assert.WithinDuration(t, order.CreatedAt, retrievedOrder.CreatedAt, time.Second)
		assert.WithinDuration(t, order.UpdatedAt, retrievedOrder.UpdatedAt, time.Second)

		// Assert order items
		assert.Len(t, retrievedOrder.Items, len(order.Items))
		// Since order of items might vary, convert to map for easier comparison
		retrievedItemsMap := make(map[uuid.UUID]domain.OrderItem)
		for _, item := range retrievedOrder.Items {
			retrievedItemsMap[item.ProductID] = item
		}
		for _, originalItem := range order.Items {
			retrievedItem, ok := retrievedItemsMap[originalItem.ProductID]
			assert.True(t, ok, "Retrieved item for product ID %s not found", originalItem.ProductID)
			assert.Equal(t, originalItem.Quantity, retrievedItem.Quantity)
			assert.InDelta(t, originalItem.UnitPrice, retrievedItem.UnitPrice, 0.001)
		}
	})

	t.Run("Get Non-Existent Order", func(t *testing.T) {
		t.Parallel()
		nonExistentID := uuid.New()
		order, err := repo.GetOrderByID(ctx, nonExistentID)
		assert.ErrorIs(t, err, domain.ErrOrderNotFound, "Expected ErrOrderNotFound for non-existent order")
		assert.Nil(t, order, "Expected nil order for non-existent ID")
	})

	t.Run("Create order with duplicate ID (simulating failure)", func(t *testing.T) {
		t.Parallel()
		customerID := uuid.New()
		productID := uuid.New()
		items := []domain.OrderItem{{ProductID: productID, Quantity: 1, UnitPrice: 1.0}}

		order1, _ := domain.NewOrder(customerID, items)
		order1.ID = uuid.New() // Ensure unique ID for this specific test case

		err := repo.CreateOrder(ctx, order1)
		assert.NoError(t, err, "Expected no error for first creation")

		// Try to create another order with the same ID (this should fail due to PK constraint)
		order2, _ := domain.NewOrder(customerID, items)
		order2.ID = order1.ID // Assign same ID
		err = repo.CreateOrder(ctx, order2)
		assert.Error(t, err, "Expected error for duplicate ID creation")
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
	})
}
