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

		// Generate realistic ecommerce metrics
		rating := randomRating()                 // 0-5 stars (weighted towards higher ratings)
		reviewCount := randomReviewCount(rating) // More reviews for popular products
		salesCount := randomSalesCount(rating)   // Sales correlate with ratings
		viewCount := randomViewCount(salesCount) // Views correlate with sales
		ctr := randomCTR(rating)                 // CTR correlates with rating
		isPromoted := rand.Float64() < 0.15      // 15% of products are promoted
		margin := randomMargin(price)            // Higher price items often have better margins

		product := &models.Product{
			Name:        name,
			Description: fmt.Sprintf("%s designed for %s use with premium build quality.", name, category),
			Price:       price,
			Category:    category,
			Stock:       stock,
			Rating:      rating,
			ReviewCount: reviewCount,
			SalesCount:  salesCount,
			ViewCount:   viewCount,
			CTR:         ctr,
			IsPromoted:  isPromoted,
			Margin:      margin,
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

// randomRating generates ratings weighted towards higher values (realistic distribution)
func randomRating() float64 {
	// Weighted distribution: more 4-5 star products than 1-2 star
	weights := []float64{0.05, 0.10, 0.15, 0.30, 0.40} // 1*, 2*, 3*, 4*, 5*
	r := rand.Float64()
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			// Return rating with decimal (e.g., 4.2, 4.7)
			base := float64(i + 1)
			decimal := rand.Float64() * 0.9 // 0.0 - 0.9
			rating := base + decimal
			if rating > 5.0 {
				rating = 5.0
			}
			return float64(int(rating*10)) / 10 // Round to 1 decimal
		}
	}
	return 4.5
}

// randomReviewCount generates review count based on rating (better products have more reviews)
func randomReviewCount(rating float64) int {
	// Higher rated products tend to have more reviews
	baseReviews := int(rating * 50) // 0-250 base
	variance := rand.Intn(200)      // 0-200 variance
	return baseReviews + variance
}

// randomSalesCount generates sales based on rating (better products sell more)
func randomSalesCount(rating float64) int {
	// Sales correlate strongly with ratings
	baseSales := int(rating * 200) // 0-1000 base
	variance := rand.Intn(500)     // 0-500 variance
	return baseSales + variance
}

// randomViewCount generates view count based on sales (viewed products get bought)
func randomViewCount(salesCount int) int {
	// Typical conversion rate is 2-5%, so views = sales * 20-50
	multiplier := 20 + rand.Intn(30) // 20-50x
	return salesCount * multiplier
}

// randomCTR generates click-through rate based on rating
func randomCTR(rating float64) float64 {
	// Higher rated products have better CTR
	// Range: 0.02 (2%) to 0.15 (15%)
	baseCTR := 0.02 + (rating/5.0)*0.13
	variance := (rand.Float64() - 0.5) * 0.03 // +/- 1.5%
	ctr := baseCTR + variance

	if ctr < 0.01 {
		ctr = 0.01
	}
	if ctr > 0.20 {
		ctr = 0.20
	}

	return float64(int(ctr*1000)) / 1000 // Round to 3 decimals
}

// randomMargin generates profit margin based on price (higher price = better margin typically)
func randomMargin(price float64) float64 {
	// Expensive items often have better margins
	// Range: 0.15 (15%) to 0.45 (45%)
	baseMargin := 0.15 + (price/2500.0)*0.20
	variance := rand.Float64() * 0.10
	margin := baseMargin + variance

	if margin > 0.45 {
		margin = 0.45
	}

	return float64(int(margin*100)) / 100 // Round to 2 decimals
}
