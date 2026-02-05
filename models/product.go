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
