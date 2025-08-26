package services

import (
	"context"
	"errors"
	"time"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

type OrderService struct { store *storage.MemoryStore }

func NewOrderService(store *storage.MemoryStore) *OrderService { return &OrderService{store: store} }

func (s *OrderService) PlaceOrder(ctx context.Context, req *dto.PlaceOrderRequest) (*dto.Order, error) {
	_ = ctx
	// Duplicate order guard
	orders, _ := s.store.GetOrdersByUser(req.UserID)
	if len(orders) > 0 {
		last := orders[len(orders)-1]
		if time.Since(last.CreatedAt) <= time.Duration(DuplicateOrderWindowSec)*time.Second {
			return nil, errors.New("duplicate order detected")
		}
	}
	// Validate cart and compute total without mutating stock/cart
	cart, err := s.store.GetCartByUser(req.UserID)
	if err != nil { return nil, err }
	if len(cart.Items) == 0 { return nil, errors.New("cart is empty") }
	// Special item alone check
	if len(cart.Items) > 1 {
		for _, it := range cart.Items {
			p, err := s.store.GetProductByID(it.ProductID)
			if err != nil { return nil, err }
			if p.IsSpecial { return nil, errors.New("special items must be purchased alone") }
		}
	}
	total := 0.0
	for _, it := range cart.Items {
		p, err := s.store.GetProductByID(it.ProductID)
		if err != nil { return nil, err }
		if it.Quantity <= 0 { return nil, errors.New("invalid cart item quantity") }
		if p.IsSpecial && it.Quantity != 1 { return nil, errors.New("special items must have quantity 1") }
		if it.UnitPrice != 0 && it.UnitPrice != p.Price { return nil, errors.New("prices changed, refresh cart") }
		if p.Stock < it.Quantity { return nil, errors.New("insufficient stock for product") }
		total += float64(it.Quantity) * p.Price
	}
	if total < MinOrderAmount { return nil, errors.New("order total below minimum") }
	// Daily spend cap
	// sum today's orders totals
	todayTotal := 0.0
	now := time.Now()
	for _, o := range orders {
		if sameDay(now, o.CreatedAt) {
			todayTotal += o.Total
		}
	}
	if todayTotal+total > DailyUserSpendCap { return nil, errors.New("daily spend limit reached") }
	// Reserve stock and create order
	order, err := s.store.ReserveStockForOrder(req.UserID)
	if err != nil { return nil, err }
	if order.Total > HighValueReviewThreshold {
		order.Status = "PENDING_REVIEW"
	}
	return s.store.CreateOrder(order)
}

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

