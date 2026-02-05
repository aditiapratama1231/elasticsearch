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

	// Text search on name and description with fuzzy matching
	if searchReq.Query != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     searchReq.Query,
				"fields":    []string{"name^2", "description"},
				"fuzziness": "AUTO",
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

	// Apply custom scoring formula: FS = a * ³√(RS * b * (SS * d)) + log10((BS * d) + 1)
	// RS=Relevance Score, SS=Stock Score, BS=Business Score, a=1.2, b=0.8, d=0.01
	scoringQuery := map[string]interface{}{
		"script_score": map[string]interface{}{
			"query": query,
			"script": map[string]interface{}{
				"source": "Math.pow(_score * 0.8 * (Math.max(doc['stock'].value, 1) * 0.01), 1.0/3.0) * 1.2 + Math.log10((Math.max(doc['stock'].value, 1) * 0.01) + 1)",
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
