package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aditya/elasticsearch-products-api/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type ProductRepository struct {
	client    *elasticsearch.Client
	indexName string
}

func NewProductRepository(client *elasticsearch.Client, indexName string) *ProductRepository {
	return &ProductRepository{
		client:    client,
		indexName: indexName,
	}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	product.ID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("error marshaling product: %w", err)
	}

	log.Printf("[ES] CREATE - Index: %s, DocumentID: %s, Body: %s", r.indexName, product.ID, string(data))

	req := esapi.IndexRequest{
		Index:      r.indexName,
		DocumentID: product.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("error indexing product: %w", err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	log.Printf("[ES] CREATE RESPONSE - Status: %d, Response: %s", res.StatusCode, string(resBody))

	if res.IsError() {
		return fmt.Errorf("error response: %s", string(resBody))
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	log.Printf("[ES] GET - Index: %s, DocumentID: %s", r.indexName, id)

	req := esapi.GetRequest{
		Index:      r.indexName,
		DocumentID: id,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("error getting product: %w", err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	log.Printf("[ES] GET RESPONSE - Status: %d, Response: %s", res.StatusCode, string(resBody))

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("error response: %s", string(resBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resBody, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	source := result["_source"].(map[string]interface{})
	productData, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("error marshaling source: %w", err)
	}

	var product models.Product
	if err := json.Unmarshal(productData, &product); err != nil {
		return nil, fmt.Errorf("error unmarshaling product: %w", err)
	}

	return &product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, id string, product *models.Product) error {
	// First check if product exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	product.ID = id
	product.UpdatedAt = time.Now()

	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("error marshaling product: %w", err)
	}

	log.Printf("[ES] UPDATE - Index: %s, DocumentID: %s, Body: %s", r.indexName, id, string(data))

	req := esapi.IndexRequest{
		Index:      r.indexName,
		DocumentID: id,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	log.Printf("[ES] UPDATE RESPONSE - Status: %d, Response: %s", res.StatusCode, string(resBody))

	if res.IsError() {
		return fmt.Errorf("error response: %s", string(resBody))
	}

	return nil
}

// Delete deletes a product by ID
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	log.Printf("[ES] DELETE - Index: %s, DocumentID: %s", r.indexName, id)

	req := esapi.DeleteRequest{
		Index:      r.indexName,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	log.Printf("[ES] DELETE RESPONSE - Status: %d, Response: %s", res.StatusCode, string(resBody))

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("error response: %s", string(resBody))
	}

	return nil
}

// Search searches for products based on criteria
func (r *ProductRepository) Search(ctx context.Context, searchReq *models.ProductSearchRequest) ([]models.Product, int, error) {
	// Set default pagination
	if searchReq.Page < 1 {
		searchReq.Page = 1
	}
	if searchReq.PageSize < 1 {
		searchReq.PageSize = 10
	}

	from := (searchReq.Page - 1) * searchReq.PageSize

	// Build query
	var query map[string]interface{}

	mustClauses := []map[string]interface{}{}

	// Text search on name and description with edge n-grams for autocomplete and fuzzy matching
	if searchReq.Query != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     searchReq.Query,
				"fields":    []string{"name.autocomplete^3", "name^2", "description.autocomplete", "description"},
				"fuzziness": "AUTO",
				"type":      "best_fields",
			},
		})
	}

	// Category filter
	if searchReq.Category != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"category": searchReq.Category,
			},
		})
	}

	// Price range filter
	if searchReq.MinPrice > 0 || searchReq.MaxPrice > 0 {
		priceRange := map[string]interface{}{}
		if searchReq.MinPrice > 0 {
			priceRange["gte"] = searchReq.MinPrice
		}
		if searchReq.MaxPrice > 0 {
			priceRange["lte"] = searchReq.MaxPrice
		}
		mustClauses = append(mustClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"price": priceRange,
			},
		})
	}

	if len(mustClauses) > 0 {
		query = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		}
	} else {
		query = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	// Apply enhanced ecommerce scoring formula
	// Components:
	// 1. Base relevance (_score from text matching)
	// 2. Stock availability (in-stock boost, out-of-stock penalty)
	// 3. Rating boost (higher rated products rank higher)
	// 4. Social proof (review count logarithmic boost)
	// 5. Popularity (sales count logarithmic boost)
	// 6. Engagement (CTR and view count)
	// 7. Business rules (promoted products, margin)
	scoringQuery := map[string]interface{}{
		"script_score": map[string]interface{}{
			"query": query,
			"script": map[string]interface{}{
				"source": `
					// Base relevance score from text matching
					double baseScore = _score;
					
					// Stock availability: out-of-stock = 0.3x penalty, in-stock = 1.0x
					double stockMultiplier = doc['stock'].value > 0 ? 1.0 : 0.3;
					
					// Rating boost: normalize 0-5 rating to 0.6-1.2 multiplier
					// (3 stars = 1.0x, 5 stars = 1.2x, 0 stars = 0.6x)
					double ratingBoost = doc['review_count'].value > 0 
						? 0.6 + (doc['rating'].value / 5.0) * 0.6 
						: 1.0;
					
					// Social proof: logarithmic boost from review count
					// More reviews = more trust (diminishing returns)
					double reviewBoost = 1.0 + Math.log10(doc['review_count'].value + 1) * 0.1;
					
					// Popularity: logarithmic boost from sales count
					// Best sellers rank higher
					double popularityBoost = 1.0 + Math.log10(doc['sales_count'].value + 1) * 0.15;
					
					// Engagement: CTR and view count combined
					// High CTR = users find it relevant
					double engagementBoost = 1.0 + (doc['ctr'].value * 0.2) + (Math.log10(doc['view_count'].value + 1) * 0.05);
					
					// Business boost: promoted products + margin consideration
					// Promoted products get 1.3x boost, high margin products get slight boost
					double businessBoost = (doc['is_promoted'].value ? 1.3 : 1.0) * (1.0 + doc['margin'].value * 0.1);
					
					// Final score: combine all signals
					return baseScore * stockMultiplier * ratingBoost * reviewBoost * popularityBoost * engagementBoost * businessBoost;
				`,
			},
		},
	}

	searchBody := map[string]interface{}{
		"query": scoringQuery,
		"from":  from,
		"size":  searchReq.PageSize,
		"sort": []map[string]interface{}{
			{"_score": map[string]interface{}{"order": "desc"}},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, 0, fmt.Errorf("error encoding search query: %w", err)
	}

	queryStr := buf.String()
	log.Printf("[ES] SEARCH - Index: %s, Query: %s", r.indexName, queryStr)

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.indexName),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	log.Printf("[ES] SEARCH RESPONSE - Status: %d, Response: %s", res.StatusCode, string(resBody))

	if res.IsError() {
		return nil, 0, fmt.Errorf("error response: %s", string(resBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resBody, &result); err != nil {
		return nil, 0, fmt.Errorf("error decoding response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := int(hits["total"].(map[string]interface{})["value"].(float64))
	hitsArray := hits["hits"].([]interface{})

	products := make([]models.Product, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		productData, err := json.Marshal(source)
		if err != nil {
			continue
		}

		var product models.Product
		if err := json.Unmarshal(productData, &product); err != nil {
			continue
		}

		products = append(products, product)
	}

	return products, total, nil
}

// GetAll retrieves all products with pagination
func (r *ProductRepository) GetAll(ctx context.Context, page, pageSize int) ([]models.Product, int, error) {
	searchReq := &models.ProductSearchRequest{
		Page:     page,
		PageSize: pageSize,
	}
	return r.Search(ctx, searchReq)
}
