package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

func TestOrderService_PlaceOrder(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	cartSvc := NewCartService(store)
	svc := NewOrderService(store)

	// add to cart for user 1
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 2}); err != nil {
		t.Fatalf("add: %v", err)
	}

	order, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1})
	if err != nil { t.Fatalf("place order: %v", err) }
	if order.UserID != 1 { t.Fatalf("expected user 1, got %d", order.UserID) }
	if len(order.Items) != 1 { t.Fatalf("expected 1 item, got %d", len(order.Items)) }
	if order.Items[0].Quantity != 2 { t.Fatalf("expected qty 2, got %d", order.Items[0].Quantity) }
	if order.Total <= 0 { t.Fatalf("expected positive total, got %f", order.Total) }

	// cart should be cleared after order
	c, err := cartSvc.GetCart(ctx, &dto.GetCartRequest{UserID: 1})
	if err != nil { t.Fatalf("get cart: %v", err) }
	if len(c.Items) != 0 { t.Fatalf("expected cart cleared after order, got %d items", len(c.Items)) }
}
