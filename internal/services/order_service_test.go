package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/endpoint"
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
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 2}}); err != nil {
		t.Fatalf("add: %v", err)
	}

	orderResp, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}})
	if err != nil { t.Fatalf("place order: %v", err) }
	order := orderResp.Body
	if order.UserID != 1 { t.Fatalf("expected user 1, got %d", order.UserID) }
	if len(order.Items) != 1 { t.Fatalf("expected 1 item, got %d", len(order.Items)) }
	if order.Items[0].Quantity != 2 { t.Fatalf("expected qty 2, got %d", order.Items[0].Quantity) }
	if order.Total <= 0 { t.Fatalf("expected positive total, got %f", order.Total) }

	// cart should be cleared after order
	cResp, err := cartSvc.GetCart(ctx, &endpoint.HTTPRequest[*dto.GetCartRequest]{Body: &dto.GetCartRequest{UserID: 1}})
	if err != nil { t.Fatalf("get cart: %v", err) }
	if len(cResp.Body.Items) != 0 { t.Fatalf("expected cart cleared after order, got %d items", len(cResp.Body.Items)) }
}

func TestOrderService_MinAmountAndHighValueReview(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	cartSvc := NewCartService(store)
	prodSvc := NewProductService(store)
	svc := NewOrderService(store)
	// min amount (price 1, qty 1)
	cheapResp, _ := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Cheap", Author: "A", Description: "", Price: 1, Stock: 10}})
	cheap := cheapResp.Body
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: cheap.ID, Quantity: 1}}); err != nil { t.Fatalf("add cheap: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}}); err == nil {
		t.Fatalf("expected min order amount error")
	}
	// high value review
	expResp, _ := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Exp", Author: "A", Description: "", Price: 2000, Stock: 10}})
	exp := expResp.Body
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: exp.ID, Quantity: 2}}); err != nil { t.Fatalf("add exp: %v", err) }
	orderResp, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}})
	order := orderResp.Body
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
	pResp, _ := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "P", Author: "A", Description: "", Price: 100, Stock: 10}})
	p := pResp.Body
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: p.ID, Quantity: 1}}); err != nil { t.Fatalf("add: %v", err) }
	// change price
	if _, err := prodSvc.UpdateProduct(ctx, &endpoint.HTTPRequest[*dto.UpdateProductRequest]{Body: &dto.UpdateProductRequest{ID: p.ID, Title: "P", Author: "A", Description: "", Price: 120, Stock: 10}}); err != nil { t.Fatalf("update: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}}); err == nil {
		t.Fatalf("expected price drift error")
	}
	// special mixed content
	specialResp, _ := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "S", Author: "A", Description: "", Price: 50, Stock: 10, IsSpecial: true}})
	special := specialResp.Body
	_, _ = cartSvc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: p.ID}})
	// add special and another product
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: special.ID, Quantity: 1}}); err != nil { t.Fatalf("add special: %v", err) }
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 2, Quantity: 1}}); err != nil { t.Fatalf("add other: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}}); err == nil {
		t.Fatalf("expected special mixed content error")
	}
	// special quantity not 1
	_, _ = cartSvc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: 2}})
	_, _ = cartSvc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: special.ID}})
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: special.ID, Quantity: 2}}); err != nil { t.Fatalf("add special 2: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}}); err == nil {
		t.Fatalf("expected special qty error")
	}
	// duplicate within 5s (even with empty cart) should be blocked
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 1}}); err == nil {
		t.Fatalf("expected duplicate order detected error")
	}
	// daily cap on another user (avoid duplicate guard)
	// two orders of 5000 then a small one exceeding cap
	bigResp, _ := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Big", Author: "A", Description: "", Price: 5000, Stock: 10}})
	big := bigResp.Body
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 2, ProductID: big.ID, Quantity: 1}}); err != nil { t.Fatalf("add big u2: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 2}}); err != nil { t.Fatalf("place big u2: %v", err) }
	// simulate second order without waiting duplicate window
	if _, err := store.CreateOrder(&models.Order{UserID: 2, Items: []models.OrderItem{}, Total: 5000, Status: "PLACED"}); err != nil {
		t.Fatalf("create direct order: %v", err)
	}
	// now any positive order should exceed daily cap
	cheapResp2, _ := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Cheap2", Author: "A", Description: "", Price: 10, Stock: 10}})
	cheap2 := cheapResp2.Body
	if _, err := cartSvc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 2, ProductID: cheap2.ID, Quantity: 1}}); err != nil { t.Fatalf("add cheap2: %v", err) }
	if _, err := svc.PlaceOrder(ctx, &endpoint.HTTPRequest[*dto.PlaceOrderRequest]{Body: &dto.PlaceOrderRequest{UserID: 2}}); err == nil {
		t.Fatalf("expected daily spend limit error")
	}
}

