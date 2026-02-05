package models

import "time"

// Product represents a product entity
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	Category    string    `json:"category" binding:"required"`
	Stock       int       `json:"stock" binding:"required,gte=0"`
	Rating      float64   `json:"rating" binding:"gte=0,lte=5"`        // 0-5 stars
	ReviewCount int       `json:"review_count" binding:"gte=0"`        // number of reviews
	SalesCount  int       `json:"sales_count" binding:"gte=0"`         // total sales
	ViewCount   int       `json:"view_count" binding:"gte=0"`          // product page views
	CTR         float64   `json:"ctr" binding:"gte=0,lte=1"`           // click-through rate (0-1)
	IsPromoted  bool      `json:"is_promoted"`                         // featured/promoted product
	Margin      float64   `json:"margin" binding:"gte=0,lte=1"`        // profit margin (0-1)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductSearchRequest represents search query parameters
type ProductSearchRequest struct {
	Query    string  `form:"q" json:"q"`
	Category string  `form:"category" json:"category"`
	MinPrice float64 `form:"min_price" json:"min_price"`
	MaxPrice float64 `form:"max_price" json:"max_price"`
	Page     int     `form:"page" json:"page"`
	PageSize int     `form:"page_size" json:"page_size"`
}
