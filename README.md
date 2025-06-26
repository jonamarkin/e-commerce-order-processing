# E-Commerce Order Processing Service

This project implements a microservice for processing customer orders as part of an e-commerce system. It handles order creation, retrieval, and publishes events to Kafka for downstream services.

## Features

* **Order Creation:** Allows customers to create new orders with multiple items.
* **Order Retrieval:** Fetch details of an order by its unique ID.
* **Event Publishing:** Publishes "Order Placed" events to a Kafka topic.
* **Persistence:** Stores order data in a PostgreSQL database.
* **API Documentation:** Interactive OpenAPI (Swagger) documentation.
* **Structured Logging:** Uses `zerolog` for clear, machine-readable logs.
* **Metrics:** Exposes Prometheus-compatible metrics for monitoring.

## Architecture

The service follows a clean architecture pattern, separating concerns into distinct layers:

* **`domain`**: Core business entities and rules (e.g., `Order`, `OrderItem`, validation logic).
* **`service`**: Business logic, orchestrating interactions between domain, repository, and external systems (Kafka).
* **`repository`**: Handles data persistence (currently PostgreSQL).
* **`kafka`**: Producer client for Kafka interactions.
* **`api`**: HTTP handlers for exposing functionality via a RESTful API.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* [Go](https://golang.org/doc/install) (1.22 or higher recommended)
* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/install/)
* [`swag` CLI tool](https://github.com/swaggo/swag): `go install github.com/swaggo/swag/cmd/swag@latest`

### Installation

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/jonamarkin/e-commerce-order-processing.git](https://github.com/jonamarkin/e-commerce-order-processing.git)
    cd e-commerce-order-processing
    ```

2.  **Set up environment variables:**
    Create a `.env` file in the project root based on `env.example`.
    ```bash
    cp env.example .env
    ```
    Edit `.env` to configure your database and Kafka settings.
    ```ini
    # .env
    PORT=8080
    DATABASE_URL="postgres://postgres:postgres@localhost:5432/ecommerce_db?sslmode=disable"
    KAFKA_BROKERS="localhost:9092"
    ```
    *Note: If running services inside Docker Compose, `localhost:9092` and `localhost:5432` refer to the host machine's exposed ports. If running from another Docker container, use service names like `kafka:9092` and `db:5432`.*

3.  **Start supporting services (PostgreSQL, Kafka, Zookeeper) with Docker Compose:**
    ```bash
    docker compose up -d db kafka zookeeper
    ```
    Wait a few moments for services to fully start. You can check their status with `docker compose ps`.

4.  **Run Database Migrations:**
    This project uses `golang-migrate`. You'll need to install it if you haven't already:
    ```bash
    go install -tags 'postgres' [github.com/golang-migrate/migrate/v4/cmd/migrate@latest](https://github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
    ```
    Apply the migrations:
    ```bash
    migrate -path database/migrations -database "$DATABASE_URL" up
    ```
    (Ensure `DATABASE_URL` is correctly set in your environment or substitute the full string.)

5.  **Generate Swagger Documentation:**
    ```bash
    swag init
    ```

6.  **Run the application:**
    ```bash
    go run cmd/orderservice/main.go
    ```
    The service will start on `http://localhost:8080`.

### API Endpoints

The API is served under `/api/v1`.

* **Interactive API Docs (Swagger UI):** `http://localhost:8080/swagger/index.html`
* **Metrics (Prometheus format):** `http://localhost:8080/metrics`

**Example cURL requests:**

* **Create Order (POST /api/v1/orders)**
    ```bash
    curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -d '{
      "customer_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef",
      "items": [
        {
          "product_id": "fedcba98-7654-3210-fedc-ba9876543210",
          "quantity": 1,
          "unit_price": 49.99
        },
        {
          "product_id": "12345678-abcd-efgh-ijkl-mnopqrstuvwx",
          "quantity": 2,
          "unit_price": 25.00
        }
      ]
    }'
    ```

* **Get Order by ID (GET /api/v1/orders/{id})**
  (Replace `<ORDER_ID>` with an ID from a created order)
    ```bash
    curl http://localhost:8080/api/v1/orders/<ORDER_ID>
    ```

### Running Tests

* **Unit Tests:**
    ```bash
    go test ./...
    ```
* **Integration Tests (Repository layer):**
    ```bash
    DATABASE_URL="postgres://postgres:postgres@localhost:5432/ecommerce_db?sslmode=disable" go test -v ./internal/orderservice/repository/...
    ```
  *Ensure your database is running and migrated before running integration tests.*

### Project Structure

```
├── cmd/               # Main application entry points
│   └── orderservice/  # Order Service main executable
├── config/            # Application configuration loading
├── database/          # Database schema migrations
│   └── migrations/
├── internal/          # Internal application code (not directly importable by other modules)
│   └── orderservice/
│       ├── api/       # HTTP handlers and API request/response models
│       ├── domain/    # Core business entities, value objects, and rules
│       ├── kafka/     # Kafka producer client
│       ├── metrics/   # Prometheus metric definitions
│       ├── repository/# Data access layer (PostgreSQL implementation)
│       └── service/   # Business logic, orchestrating domain, repo, and external calls
├── docs/              # Generated Swagger documentation
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
├── README.md          # This file
└── env.example        # Example environment variables
```



## Contributing

Feel free to open issues or submit pull requests.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.