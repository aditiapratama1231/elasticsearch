package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewElasticsearchClient(url string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error: %s", res.String())
	}

	log.Println("Successfully connected to Elasticsearch")
	return client, nil
}

func CreateProductIndex(client *elasticsearch.Client, indexName string) error {
	ctx := context.Background()

	// Check if index already exists
	exists, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}
	defer exists.Body.Close()

	if exists.StatusCode == 200 {
		log.Printf("Index '%s' already exists\n", indexName)
		return nil
	}

	// Define index mapping
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"autocomplete": map[string]interface{}{
						"tokenizer": "autocomplete_tokenizer",
						"filter":    []string{"lowercase"},
					},
					"autocomplete_search": map[string]interface{}{
						"tokenizer": "lowercase",
					},
				},
				"tokenizer": map[string]interface{}{
					"autocomplete_tokenizer": map[string]interface{}{
						"type":        "edge_ngram",
						"min_gram":    3,
						"max_gram":    15,
						"token_chars": []string{"letter", "digit"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
						"autocomplete": map[string]interface{}{
							"type":            "text",
							"analyzer":        "autocomplete",
							"search_analyzer": "autocomplete_search",
						},
					},
				},
				"description": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"autocomplete": map[string]interface{}{
							"type":            "text",
							"analyzer":        "autocomplete",
							"search_analyzer": "autocomplete_search",
						},
					},
				},
				"price": map[string]interface{}{
					"type": "float",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"stock": map[string]interface{}{
					"type": "integer",
				},
				"rating": map[string]interface{}{
					"type": "float",
				},
				"review_count": map[string]interface{}{
					"type": "integer",
				},
				"sales_count": map[string]interface{}{
					"type": "integer",
				},
				"view_count": map[string]interface{}{
					"type": "integer",
				},
				"ctr": map[string]interface{}{
					"type": "float",
				},
				"is_promoted": map[string]interface{}{
					"type": "boolean",
				},
				"margin": map[string]interface{}{
					"type": "float",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("error marshaling mapping: %w", err)
	}

	res, err := client.Indices.Create(
		indexName,
		client.Indices.Create.WithContext(ctx),
		client.Indices.Create.WithBody(bytes.NewReader(mappingJSON)),
	)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error: %s", res.String())
	}

	log.Printf("Index '%s' created successfully\n", indexName)
	return nil
}
