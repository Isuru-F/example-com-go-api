package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/models"
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

func TestOrderService_MinAmountAndHighValueReview(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	cartSvc := NewCartService(store)
	prodSvc := NewProductService(store)
	svc := NewOrderService(store)
	// min amount (price 1, qty 1)
	cheap, _ := prodSvc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "Cheap", Author: "A", Description: "", Price: 1, Stock: 10})
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: cheap.ID, Quantity: 1}); err != nil { t.Fatalf("add cheap: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1}); err == nil {
		t.Fatalf("expected min order amount error")
	}
	// high value review
	exp, _ := prodSvc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "Exp", Author: "A", Description: "", Price: 2000, Stock: 10})
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: exp.ID, Quantity: 2}); err != nil { t.Fatalf("add exp: %v", err) }
	order, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1})
	if err != nil { t.Fatalf("place: %v", err) }
	if order.Status != "PENDING_REVIEW" { t.Fatalf("expected PENDING_REVIEW, got %s", order.Status) }
}

func TestOrderService_DuplicateAndPriceDriftAndSpecialAndDailyCap(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	cartSvc := NewCartService(store)
	prodSvc := NewProductService(store)
	svc := NewOrderService(store)
	// price drift
	p, _ := prodSvc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "P", Author: "A", Description: "", Price: 100, Stock: 10})
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: p.ID, Quantity: 1}); err != nil { t.Fatalf("add: %v", err) }
	// change price
	if _, err := prodSvc.UpdateProduct(ctx, &dto.UpdateProductRequest{ID: p.ID, Title: "P", Author: "A", Description: "", Price: 120, Stock: 10}); err != nil { t.Fatalf("update: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1}); err == nil {
		t.Fatalf("expected price drift error")
	}
	// special mixed content
	special, _ := prodSvc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "S", Author: "A", Description: "", Price: 50, Stock: 10, IsSpecial: true})
	_, _ = cartSvc.RemoveFromCart(ctx, &dto.RemoveFromCartRequest{UserID: 1, ProductID: p.ID})
	// add special and another product
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: special.ID, Quantity: 1}); err != nil { t.Fatalf("add special: %v", err) }
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: 2, Quantity: 1}); err != nil { t.Fatalf("add other: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1}); err == nil {
		t.Fatalf("expected special mixed content error")
	}
	// special quantity not 1
	_, _ = cartSvc.RemoveFromCart(ctx, &dto.RemoveFromCartRequest{UserID: 1, ProductID: 2})
	_, _ = cartSvc.RemoveFromCart(ctx, &dto.RemoveFromCartRequest{UserID: 1, ProductID: special.ID})
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 1, ProductID: special.ID, Quantity: 2}); err != nil { t.Fatalf("add special 2: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1}); err == nil {
		t.Fatalf("expected special qty error")
	}
	// duplicate within 5s (even with empty cart) should be blocked
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 1}); err == nil {
		t.Fatalf("expected duplicate order detected error")
	}
	// daily cap on another user (avoid duplicate guard)
	// two orders of 5000 then a small one exceeding cap
	big, _ := prodSvc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "Big", Author: "A", Description: "", Price: 5000, Stock: 10})
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 2, ProductID: big.ID, Quantity: 1}); err != nil { t.Fatalf("add big u2: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 2}); err != nil { t.Fatalf("place big u2: %v", err) }
	// simulate second order without waiting duplicate window
	if _, err := store.CreateOrder(&models.Order{UserID: 2, Items: []models.OrderItem{}, Total: 5000, Status: "PLACED"}); err != nil {
		t.Fatalf("create direct order: %v", err)
	}
	// now any positive order should exceed daily cap
	cheap, _ := prodSvc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "Cheap2", Author: "A", Description: "", Price: 10, Stock: 10})
	if _, err := cartSvc.AddToCart(ctx, &dto.AddToCartRequest{UserID: 2, ProductID: cheap.ID, Quantity: 1}); err != nil { t.Fatalf("add cheap2: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &dto.PlaceOrderRequest{UserID: 2}); err == nil {
		t.Fatalf("expected daily spend limit error")
	}
}

