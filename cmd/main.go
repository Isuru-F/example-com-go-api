package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/handlers"
	"ecom-book-store-sample-api/internal/services"
	"ecom-book-store-sample-api/internal/storage"
)

func main() {
	store := storage.NewMemoryStore()
	storage.Seed(store)

	productSvc := services.NewProductService(store)
	cartSvc := services.NewCartService(store)
	orderSvc := services.NewOrderService(store)

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		ph := handlers.NewProductHandler(productSvc)
		api.GET("/products", ph.ListProducts)
		api.POST("/products", ph.CreateProduct)
		api.GET("/products/:id", ph.GetProduct)
		api.PUT("/products/:id", ph.UpdateProduct)
		api.DELETE("/products/:id", ph.DeleteProduct)

		ch := handlers.NewCartHandler(cartSvc)
		api.POST("/cart/user/:id/items", ch.AddToCart)
		api.DELETE("/cart/user/:id/items", ch.RemoveFromCart)

		oh := handlers.NewOrderHandler(orderSvc)
		api.POST("/orders/user/:id", oh.PlaceOrder)
	}

	srv := &http.Server{Addr: ":8080", Handler: r, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, MaxHeaderBytes: 1 << 20}

	log.Println("ecom-book-store-sample-api listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
