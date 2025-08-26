package dto

import (
	"ecom-book-store-sample-api/internal/models"
)

// Product DTOs

type CreateProductRequest struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type UpdateProductRequest struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type GetProductRequest struct { ID uint `json:"id"` }

type DeleteProductRequest struct { ID uint `json:"id"` }

type ListProductsRequest struct{}

// Cart DTOs

type AddToCartRequest struct {
	UserID    uint `json:"userId"`
	ProductID uint `json:"productId"`
	Quantity  int  `json:"quantity"`
}

type RemoveFromCartRequest struct {
	UserID    uint `json:"userId"`
	ProductID uint `json:"productId"`
}

type GetCartRequest struct { UserID uint `json:"userId"` }

// Order DTOs

type PlaceOrderRequest struct { UserID uint `json:"userId"` }

// Response aliases (1.9+ type aliases, valid in Go 1.10)

type Product = models.Product

type Cart = models.Cart

type Order = models.Order
