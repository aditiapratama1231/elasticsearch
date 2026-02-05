package main

import (
	"fmt"
	"log"

	"github.com/aditya/elasticsearch-products-api/config"
	"github.com/aditya/elasticsearch-products-api/handlers"
	"github.com/aditya/elasticsearch-products-api/repository"
	"github.com/aditya/elasticsearch-products-api/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Elasticsearch client
	esClient, err := config.NewElasticsearchClient(cfg.ElasticsearchURL)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Create index if it doesn't exist
	if err := config.CreateProductIndex(esClient, cfg.ElasticsearchIndex); err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	// Initialize repository and handler
	productRepo := repository.NewProductRepository(esClient, cfg.ElasticsearchIndex)
	productHandler := handlers.NewProductHandler(productRepo)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, productHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
