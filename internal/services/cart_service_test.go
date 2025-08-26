package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

func TestCartService_AddRemoveGet(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	svc := NewCartService(store)

	// add
	cart, err := svc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 2})
	if err != nil { t.Fatalf("add: %v", err) }
	if len(cart.Items) != 1 { t.Fatalf("expected 1 item, got %d", len(cart.Items)) }
	if cart.Items[0].Quantity != 2 { t.Fatalf("expected qty 2, got %d", cart.Items[0].Quantity) }

	// get
	got, err := svc.GetCart(ctx, &dto.GetCartRequest{UserID: 1})
	if err != nil { t.Fatalf("get cart: %v", err) }
	if len(got.Items) != 1 { t.Fatalf("expected 1 item in cart, got %d", len(got.Items)) }

	// remove
	cart, err = svc.RemoveFromCart(ctx, &dto.RemoveFromCartRequest{UserID: 1, ProductID: 1})
	if err != nil { t.Fatalf("remove: %v", err) }
	if len(cart.Items) != 0 { t.Fatalf("expected 0 items after remove, got %d", len(cart.Items)) }

	// get new user (no cart yet)
	empty, err := svc.GetCart(ctx, &dto.GetCartRequest{UserID: 999})
	if err != nil { t.Fatalf("get cart (new user): %v", err) }
	if len(empty.Items) != 0 { t.Fatalf("expected empty cart for new user, got %d", len(empty.Items)) }
}
