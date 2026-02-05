package routes

import (
	"github.com/aditya/elasticsearch-products-api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handler *handlers.ProductHandler) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", handler.CreateProduct)
			products.GET("", handler.GetAllProducts)
			products.GET("/search", handler.SearchProducts)
			products.GET("/:id", handler.GetProduct)
			products.PUT("/:id", handler.UpdateProduct)
			products.DELETE("/:id", handler.DeleteProduct)
		}
	}
}
