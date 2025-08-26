package services

import (
	"ecom-book-store-sample-api/internal/models"
	"ecom-book-store-sample-api/internal/storage"
)

type OrderService struct { store *storage.MemoryStore }

func NewOrderService(store *storage.MemoryStore) *OrderService { return &OrderService{store: store} }

func (s *OrderService) PlaceOrder(userID uint) (*models.Order, error) {
	order, err := s.store.ReserveStockForOrder(userID)
	if err != nil { return nil, err }
	return s.store.CreateOrder(order)
}
