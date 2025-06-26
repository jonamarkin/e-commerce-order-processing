package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/service"
)

// CreateOrderRequest @Description Request payload for creating a new order.
type CreateOrderRequest struct {
	CustomerID uuid.UUID         `json:"customer_id" binding:"required" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	Items      []CreateOrderItem `json:"items" binding:"required,min=1"`
}

// CreateOrderItem @Description An item within an order creation request.
type CreateOrderItem struct {
	ProductID uuid.UUID `json:"product_id" binding:"required" example:"fedcba98-7654-3210-fedc-ba9876543210"`
	Quantity  int       `json:"quantity" binding:"required,gt=0" example:"1"`
	UnitPrice float64   `json:"unit_price" binding:"required,gt=0" example:"99.99"`
}

// OrderResponse @Description Response structure for a single order.
type OrderResponse struct {
	ID         uuid.UUID           `json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	CustomerID uuid.UUID           `json:"customer_id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	Items      []OrderItemResponse `json:"items"`
	Status     string              `json:"status" example:"pending"` // Changed to string for JSON serialization
	TotalPrice float64             `json:"total_price" example:"199.98"`
	CreatedAt  time.Time           `json:"created_at" example:"2023-10-27T10:00:00Z"`
	UpdatedAt  time.Time           `json:"updated_at" example:"2023-10-27T10:00:00Z"`
}

// OrderItemResponse @Description An item within an order response.
type OrderItemResponse struct {
	ProductID uuid.UUID `json:"product_id" example:"fedcba98-7654-3210-fedc-ba9876543210"`
	Quantity  int       `json:"quantity" example:"1"`
	UnitPrice float64   `json:"unit_price" example:"99.99"`
}

// NewOrderResponse converts a domain.Order to an OrderResponse.
func NewOrderResponse(order *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}
	return OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Items:      items,
		Status:     string(order.Status), // Convert domain.OrderStatus back to string for JSON
		TotalPrice: order.TotalPrice,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

// ErrorResponse @Description Generic error response.
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request payload"`
}

// Handler holds the dependencies for our API handlers.
type Handler struct {
	orderService service.OrderService
}

// NewHandler creates a new Handler with the given OrderService.
func NewHandler(orderService service.OrderService) *Handler {
	return &Handler{
		orderService: orderService,
	}
}

// HealthCheck godoc
// @Summary Health check
// @Description Checks if the service is up and running.
// @Tags health
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// CreateOrder
// @Summary Create a new order
// @Description Create a new customer order with provided items.
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order creation request"
// @Success 201 {object} OrderResponse "Order created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// Basic validation for request data
	if req.CustomerID == uuid.Nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Customer ID is required"})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Order must contain at least one item"})
		return
	}
	for _, item := range req.Items {
		if item.ProductID == uuid.Nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Product ID is required for all items"})
			return
		}
		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Item quantity must be positive"})
			return
		}
		if item.UnitPrice <= 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Item unit price must be positive"})
			return
		}
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			UnitPrice: itemReq.UnitPrice,
		}
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), req.CustomerID, items)
	if err != nil {
		// Specific error handling for domain/service errors
		if errors.Is(err, domain.ErrNoOrderItems) ||
			errors.Is(err, domain.ErrInvalidOrderItemQuantity) ||
			errors.Is(err, domain.ErrInvalidOrderItemUnitPrice) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, NewOrderResponse(order))
}

// GetOrderByID
// @Summary Get order by ID
// @Description Get a single order's details by its unique ID.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID" Format(uuid)
// @Success 200 {object} OrderResponse "Order retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid order ID format"
// @Failure 404 {object} ErrorResponse "Order not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /orders/{id} [get]
func (h *Handler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID format"})
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get order"})
		return
	}

	c.JSON(http.StatusOK, NewOrderResponse(order))
}
