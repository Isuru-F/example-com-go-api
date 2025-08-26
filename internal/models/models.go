package models

import "time"

type User struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Product struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CartItem struct {
	ProductID uint `json:"productId"`
	Quantity  int  `json:"quantity"`
}

type Cart struct {
	ID     uint       `json:"id"`
	UserID uint       `json:"userId"`
	Items  []CartItem `json:"items"`
}

type OrderItem struct {
	ProductID uint    `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Subtotal  float64 `json:"subtotal"`
}

type Order struct {
	ID        uint        `json:"id"`
	UserID    uint        `json:"userId"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
}
