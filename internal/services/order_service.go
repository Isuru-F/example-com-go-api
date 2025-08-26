package services

import (
	"context"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

type OrderService struct { store *storage.MemoryStore }

func NewOrderService(store *storage.MemoryStore) *OrderService { return &OrderService{store: store} }

func (s *OrderService) PlaceOrder(ctx context.Context, req *dto.PlaceOrderRequest) (*dto.Order, error) {
	_ = ctx
	order, err := s.store.ReserveStockForOrder(req.UserID)
	if err != nil { return nil, err }
	return s.store.CreateOrder(order)
}
