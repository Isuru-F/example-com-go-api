package services

import (
	"context"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

type CartService struct { store *storage.MemoryStore }

func NewCartService(store *storage.MemoryStore) *CartService { return &CartService{store: store} }

func (s *CartService) AddToCart(ctx context.Context, req *dto.AddToCartRequest) (*dto.Cart, error) {
	_ = ctx
	return s.store.AddToCart(req.UserID, req.ProductID, req.Quantity)
}

func (s *CartService) RemoveFromCart(ctx context.Context, req *dto.RemoveFromCartRequest) (*dto.Cart, error) {
	_ = ctx
	return s.store.RemoveFromCart(req.UserID, req.ProductID)
}

func (s *CartService) GetCart(ctx context.Context, req *dto.GetCartRequest) (*dto.Cart, error) {
	_ = ctx
	return s.store.GetCartByUser(req.UserID)
}
