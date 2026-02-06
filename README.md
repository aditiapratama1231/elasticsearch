# Elasticsearch Products API

A production-ready RESTful API built with Go and Elasticsearch for managing products with full CRUD operations and **advanced ecommerce search capabilities** including autocomplete, fuzzy matching, and intelligent ranking.

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Search Engine**: Elasticsearch 8.x
- **Client Library**: elastic/go-elasticsearch (official)
- **Configuration**: godotenv

## Features

### Core Functionality
- ✅ Create, Read, Update, Delete (CRUD) operations for products
- ✅ Full-text search on product name and description
- ✅ Filter by category and price range
- ✅ Pagination support
- ✅ RESTful API design
- ✅ Elasticsearch integration with proper indexing

### Advanced Search Features
- 🔍 **Autocomplete**: Edge n-gram tokenization (3-15 characters) for prefix matching
- 🎯 **Fuzzy Matching**: Typo-tolerant search with AUTO fuzziness
- ⭐ **Intelligent Ranking**: Multi-signal scoring algorithm considering:
  - Text relevance (BM25)
  - Product ratings & reviews
  - Sales velocity & popularity
  - User engagement (CTR, views)
  - Stock availability
  - Business rules (promoted products, margins)
- 📊 **Field Boosting**: Prioritizes matches in product names over descriptions

## Scoring Algorithm

The search ranking uses a **7-factor ecommerce scoring formula** that combines text relevance with business metrics:

### Mathematical Formula

```
FRS = BS × S × R × Re × P × E × B
```

**Where:**

| Component | Formula | Description |
|-----------|---------|-------------|
| **Base Score (BS)** | `BS = _score` | Elasticsearch BM25 text relevance score |
| **Stock Multiplier (S)** | `S = stock > 0 ? 1.0 : 0.3` | Out-of-stock penalty (70% reduction) |
| **Rating Boost (R)** | `R = 0.6 + (rating/5.0) × 0.6`<br/>*if review_count > 0* | Range: 0.6× (0★) to 1.2× (5★)<br/>Neutral at 3★ = 1.0× |
| **Review Boost (Re)** | `Re = 1.0 + log₁₀(review_count + 1) × 0.1` | Social proof with logarithmic scaling |
| **Popularity Boost (P)** | `P = 1.0 + log₁₀(sales_count + 1) × 0.15` | Best sellers rank higher |
| **Engagement Boost (E)** | `E = 1.0 + (ctr × 0.2) + (log₁₀(view_count + 1) × 0.05)` | CTR + view count combined |
| **Business Boost (B)** | `B = (promoted ? 1.3 : 1.0) × (1.0 + margin × 0.1)` | Promoted products get 30% boost + margin |

### Scoring Example

Product: **"Gaming Laptop"** with the following attributes:

| Attribute | Value |
|-----------|-------|
| Text match score | 2.5 |
| Stock | 50 (in stock) |
| Rating | 5.0★ (300 reviews) |
| Sales | 1000 units |
| Views | 40,000 |
| CTR | 0.15 (15%) |
| Promoted | Yes |
| Margin | 0.30 (30%) |

**Step-by-step calculation:**

```
BS  = 2.5                                    (text relevance)
S   = 1.0                                    (in stock)
R   = 0.6 + (5.0/5.0) × 0.6 = 1.2           (5★ rating)
Re  = 1.0 + log₁₀(301) × 0.1 = 1.248        (300 reviews)
P   = 1.0 + log₁₀(1001) × 0.15 = 1.45       (1000 sales)
E   = 1.0 + (0.15 × 0.2) + (log₁₀(40001) × 0.05) = 1.26
B   = 1.3 × (1.0 + 0.30 × 0.1) = 1.339      (promoted + 30% margin)

FRS = 2.5 × 1.0 × 1.2 × 1.248 × 1.45 × 1.26 × 1.339 = 8.95
```

**Result:** This product ranks **8.95×** higher than its base text relevance score!

### Field Boosting

Multi-match query uses the following field weights:
- `name.autocomplete^3` - Highest priority (3x boost)
- `name^2` - Standard name field (2x boost)
- `description.autocomplete` - Description prefix matches
- `description` - Standard description field

This ensures product names are prioritized over descriptions in search results.

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

### Search Not Finding Results
- Ensure products are indexed: `curl http://localhost:9200/products/_count`
- Check minimum query length (min_gram = 3, requires at least 3 characters)
- Verify field mappings: `curl http://localhost:9200/products/_mapping`

### Score Too Low/High
- Adjust coefficients in `repository/product_repository.go` scoring formula
- Common adjustments:
  - `ratingBoost`: Change `0.6 + (rating/5.0) * 0.6` multipliers
  - `reviewBoost`: Adjust `0.1` coefficient for review impact
  - `popularityBoost`: Adjust `0.15` coefficient for sales impact
  - `businessBoost`: Change `1.3` for promoted product boost

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
  "stock": 50,
  "rating": 4.5,
  "review_count": 250,
  "sales_count": 1200,
  "view_count": 35000,
  "ctr": 0.12,
  "is_promoted": false,
  "margin": 0.25
}
```

**Product Model Fields:**
- `name` (required): Product name
- `description`: Product description
- `price` (required): Price in USD
- `category` (required): Product category
- `stock` (required): Available quantity
- `rating`: Star rating (0-5)
- `review_count`: Number of reviews
- `sales_count`: Total sales
- `view_count`: Product page views
- `ctr`: Click-through rate (0-1)
- `is_promoted`: Featured/promoted flag
- `margin`: Profit margin (0-1)

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
- `q`: Search query (searches in name and description with autocomplete & fuzzy matching)
- `category`: Filter by category
- `min_price`: Minimum price
- `max_price`: Maximum price
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10)

**Search Examples:**
```bash
# Autocomplete: "lap" matches "Laptop"
curl "http://localhost:8080/api/v1/products/search?q=lap"

# Fuzzy: "lptop" matches "Laptop" (typo tolerance)
curl "http://localhost:8080/api/v1/products/search?q=lptop"

# Multi-word: "gaming laptop"
curl "http://localhost:8080/api/v1/products/search?q=gaming%20laptop"

# With filters
curl "http://localhost:8080/api/v1/products/search?q=laptop&category=electronics&min_price=500&max_price=2000"
```

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

## Seeding Test Data

Generate 100 realistic ecommerce products with varied ratings, reviews, and sales data:

```bash
go run cmd/seed/main.go
```

The seeder creates products with:
- Random combinations of adjectives and product types
- Weighted rating distribution (more 4-5★ products)
- Correlated metrics (high ratings → more reviews/sales)
- Realistic CTR values (2-15%)
- 15% of products marked as promoted
- Variable profit margins (15-45%)

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

### Standard Fields
- `id`: keyword
- `name`: text with keyword and autocomplete sub-fields
- `description`: text with autocomplete sub-field
- `price`: float
- `category`: keyword
- `stock`: integer
- `created_at`: date
- `updated_at`: date

### Ecommerce Ranking Fields
- `rating`: float (0-5 stars)
- `review_count`: integer (social proof)
- `sales_count`: integer (popularity)
- `view_count`: integer (engagement)
- `ctr`: float (click-through rate, 0-1)
- `is_promoted`: boolean (business rule)
- `margin`: float (profitability, 0-1)

### Text Analyzers

**Autocomplete Analyzer** (indexing):
- Tokenizer: edge_ngram (min_gram: 3, max_gram: 15)
- Filter: lowercase
- Purpose: Enables prefix matching ("lap" → "Laptop")

**Autocomplete Search Analyzer** (searching):
- Tokenizer: lowercase
- Purpose: Prevents double n-gramming at search time

## Performance Considerations

- **Autocomplete**: Edge n-grams increase index size by ~30-50%
- **Fuzzy matching**: Adds query latency (~10-20ms per query)
- **Script scoring**: More expensive than standard relevance scoring
- **Recommended**: Add caching layer (Redis) for popular queries
- **Scaling**: Consider multiple Elasticsearch nodes for >100K products

## Future Enhancements

- [ ] Add synonym support (e.g., "phone" → "mobile", "smartphone")
- [ ] Implement query suggestions (did you mean?)
- [ ] Add faceted search (price ranges, rating buckets)
- [ ] Real-time inventory updates via Elasticsearch update API
- [ ] A/B testing framework for scoring formula optimization
- [ ] Machine learning rank learning (LTR)
- [ ] Personalized search based on user history
- [ ] Multi-language support

## License

MIT License
