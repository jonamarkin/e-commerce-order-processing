package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/domain"
	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/service"
)

type CreateOrderRequest struct {
	CustomerID string             `json:"customer_id" binding:"required,uuid"`
	Items      []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type OrderItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"`
}

type OrderResponse struct {
	ID         uuid.UUID           `json:"id"`
	CustomerID uuid.UUID           `json:"customer_id"`
	Items      []OrderItemResponse `json:"items"`
	Status     domain.OrderStatus  `json:"status"`
	TotalPrice float64             `json:"total_price"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
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
		Status:     domain.OrderStatus(order.Status),
		TotalPrice: order.TotalPrice,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
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

// CreateOrder godoc
// @Summary Create a new order
// @Description Creates a new customer order with specified items.
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Create order request"
// @Success 201 {object} OrderResponse "Order created successfully"
// @Failure 400 {object} map[string]string "Bad request, invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	// Bind JSON request body to our struct and validate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
		return
	}

	domainItems := make([]domain.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format in item"})
			return
		}
		domainItems[i] = domain.OrderItem{
			ProductID: productID,
			Quantity:  itemReq.Quantity,
			UnitPrice: itemReq.UnitPrice,
		}
	}

	// Call the service layer with a context that has a timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := h.orderService.CreateOrder(ctx, customerID, domainItems)
	if err != nil {
		// Map service errors to appropriate HTTP responses
		if errors.Is(err, domain.ErrNoOrderItems) ||
			errors.Is(err, domain.ErrInvalidOrderItemQuantity) ||
			errors.Is(err, domain.ErrInvalidOrderItemUnitPrice) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, NewOrderResponse(order))
}

// GetOrderByID godoc
// @Summary Get an order by ID
// @Description Retrieve a single order by its unique ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} OrderResponse "Order retrieved successfully"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders/{id} [get]
func (h *Handler) GetOrderByID(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	// Call the service layer with a context that has a timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, NewOrderResponse(order))
}
