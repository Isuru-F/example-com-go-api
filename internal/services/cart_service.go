package services

import (
	"ecom-book-store-sample-api/internal/models"
	"ecom-book-store-sample-api/internal/storage"
)

type CartService struct {
	store *storage.MemoryStore
}

func NewCartService(store *storage.MemoryStore) *CartService { return &CartService{store: store} }

func (s *CartService) AddToCart(userID, productID uint, quantity int) (*models.Cart, error) {
	return s.store.AddToCart(userID, productID, quantity)
}

func (s *CartService) RemoveFromCart(userID, productID uint) (*models.Cart, error) {
	return s.store.RemoveFromCart(userID, productID)
}

func (s *CartService) GetCart(userID uint) (*models.Cart, error) { return s.store.GetCartByUser(userID) }
