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
					},
				},
				"description": map[string]interface{}{
					"type": "text",
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
