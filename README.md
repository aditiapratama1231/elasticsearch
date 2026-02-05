# Elasticsearch Products API

A RESTful API built with Go and Elasticsearch for managing products with full CRUD operations and advanced search capabilities.

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Search Engine**: Elasticsearch 8.x
- **Client Library**: elastic/go-elasticsearch (official)
- **Configuration**: godotenv

## Features

- ✅ Create, Read, Update, Delete (CRUD) operations for products
- ✅ Full-text search on product name and description
- ✅ Filter by category and price range
- ✅ Pagination support
- ✅ RESTful API design
- ✅ Elasticsearch integration with proper indexing

## Project Structure

```
.
├── config/              # Configuration and Elasticsearch setup
├── models/              # Data models
├── repository/          # Data access layer
├── handlers/            # HTTP handlers
├── routes/              # Route definitions
├── main.go             # Application entry point
├── docker-compose.yml  # Docker setup for Elasticsearch
└── .env.example        # Environment variables template
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for running Elasticsearch)

## Setup Instructions

### 1. Clone the Repository

```bash
cd /path/to/elasticsearch
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` if needed (default values should work for local development).

### 4. Start Elasticsearch

```bash
docker-compose up -d
```

Wait a few seconds for Elasticsearch to be ready. Check status:

```bash
curl http://localhost:9200
```

### 5. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

### Health Check
```bash
GET /health
```

### Create Product
```bash
POST /api/v1/products
Content-Type: application/json

{
  "name": "Laptop",
  "description": "High-performance laptop for developers",
  "price": 1299.99,
  "category": "electronics",
  "stock": 50
}
```

### Get All Products
```bash
GET /api/v1/products?page=1&page_size=10
```

### Get Product by ID
```bash
GET /api/v1/products/{id}
```

### Update Product
```bash
PUT /api/v1/products/{id}
Content-Type: application/json

{
  "name": "Gaming Laptop",
  "description": "High-end gaming laptop",
  "price": 1499.99,
  "category": "electronics",
  "stock": 30
}
```

### Delete Product
```bash
DELETE /api/v1/products/{id}
```

### Search Products
```bash
GET /api/v1/products/search?q=laptop&category=electronics&min_price=1000&max_price=2000&page=1&page_size=10
```

Query parameters:
- `q`: Search query (searches in name and description)
- `category`: Filter by category
- `min_price`: Minimum price
- `max_price`: Maximum price
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10)

## Example Usage

### Create a Product
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro",
    "description": "Apple MacBook Pro 16-inch with M2 chip",
    "price": 2499.99,
    "category": "electronics",
    "stock": 25
  }'
```

### Search Products
```bash
curl "http://localhost:8080/api/v1/products/search?q=macbook&category=electronics"
```

### Get All Products
```bash
curl "http://localhost:8080/api/v1/products?page=1&page_size=10"
```

## Development

### Run Tests
```bash
go test ./...
```

### Build Binary
```bash
go build -o bin/api main.go
```

### Run Binary
```bash
./bin/api
```

## Elasticsearch Index Mapping

The products index uses the following mapping:

- `id`: keyword
- `name`: text with keyword field
- `description`: text
- `price`: float
- `category`: keyword
- `stock`: integer
- `created_at`: date
- `updated_at`: date

## Stopping the Application

1. Stop the Go application: `Ctrl+C`
2. Stop Elasticsearch:
```bash
docker-compose down
```

To remove data volumes:
```bash
docker-compose down -v
```

## Troubleshooting

### Elasticsearch Connection Error
- Ensure Docker is running
- Verify Elasticsearch is up: `docker-compose ps`
- Check logs: `docker-compose logs elasticsearch`

### Port Already in Use
- Change `SERVER_PORT` in `.env`
- Or stop the process using port 8080

## License

MIT License
