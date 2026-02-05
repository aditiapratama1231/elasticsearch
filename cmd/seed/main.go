package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/aditya/elasticsearch-products-api/config"
	"github.com/aditya/elasticsearch-products-api/models"
	"github.com/aditya/elasticsearch-products-api/repository"
)

func main() {
	cfg := config.LoadConfig()

	esClient, err := config.NewElasticsearchClient(cfg.ElasticsearchURL)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	if err := config.CreateProductIndex(esClient, cfg.ElasticsearchIndex); err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	repo := repository.NewProductRepository(esClient, cfg.ElasticsearchIndex)

	seedProducts(repo, 100)
}

func seedProducts(repo *repository.ProductRepository, count int) {
	adjectives := []string{"Ultra", "Pro", "Gaming", "Smart", "Portable", "Compact", "Premium", "Eco", "Wireless", "Classic", "Advanced", "Budget", "Rugged", "Lightweight", "High-End"}
	nouns := []string{"Laptop", "Headphones", "Keyboard", "Mouse", "Monitor", "Phone", "Tablet", "Camera", "Speaker", "Router", "Backpack", "Chair", "Desk", "Microphone", "Smartwatch"}
	categories := []string{"electronics", "accessories", "office", "audio", "gaming"}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	ctx := context.Background()
	created := 0

	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s %s", adjectives[rand.Intn(len(adjectives))], nouns[rand.Intn(len(nouns))])
		category := categories[rand.Intn(len(categories))]
		price := randomPrice(49.99, 2499.99)
		stock := rand.Intn(200)

		product := &models.Product{
			Name:        name,
			Description: fmt.Sprintf("%s designed for %s use with premium build quality.", name, category),
			Price:       price,
			Category:    category,
			Stock:       stock,
		}

		if err := repo.Create(ctx, product); err != nil {
			log.Printf("Failed to create product %d: %v", i+1, err)
			continue
		}
		created++
	}

	log.Printf("Seed complete. Created %d/%d products", created, count)
}

func randomPrice(min, max float64) float64 {
	value := min + rand.Float64()*(max-min)
	return float64(int(value*100)) / 100
}
