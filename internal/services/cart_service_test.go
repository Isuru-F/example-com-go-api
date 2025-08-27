package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/endpoint"
	"ecom-book-store-sample-api/internal/storage"
)

func TestCartService_AddRemoveGet(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	svc := NewCartService(store)

	// add
	cartResp, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 2}})
	if err != nil { t.Fatalf("add: %v", err) }
	cart := cartResp.Body
	if len(cart.Items) != 1 { t.Fatalf("expected 1 item, got %d", len(cart.Items)) }
	if cart.Items[0].Quantity != 2 { t.Fatalf("expected qty 2, got %d", cart.Items[0].Quantity) }

	// get
	gotResp, err := svc.GetCart(ctx, &endpoint.HTTPRequest[*dto.GetCartRequest]{Body: &dto.GetCartRequest{UserID: 1}})
	if err != nil { t.Fatalf("get cart: %v", err) }
	got := gotResp.Body
	if len(got.Items) != 1 { t.Fatalf("expected 1 item in cart, got %d", len(got.Items)) }

	// remove
	cartResp, err = svc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: 1}})
	if err != nil { t.Fatalf("remove: %v", err) }
	cart = cartResp.Body
	if len(cart.Items) != 0 { t.Fatalf("expected 0 items after remove, got %d", len(cart.Items)) }

	// get new user (no cart yet)
	emptyResp, err := svc.GetCart(ctx, &endpoint.HTTPRequest[*dto.GetCartRequest]{Body: &dto.GetCartRequest{UserID: 999}})
	if err != nil { t.Fatalf("get cart (new user): %v", err) }
	empty := emptyResp.Body
	if len(empty.Items) != 0 { t.Fatalf("expected empty cart for new user, got %d", len(empty.Items)) }
}

func TestCartService_Rules(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	prodSvc := NewProductService(store)
	svc := NewCartService(store)
	// create an expensive product for risk limit
	expResp, err := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Expensive", Author: "A", Description: "", Price: 2000, Stock: 10}})
	if err != nil { t.Fatalf("create expensive: %v", err) }
	exp := expResp.Body
	// risk under limit ok (2*2000=4000)
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: exp.ID, Quantity: 2}}); err != nil { t.Fatalf("risk under limit: %v", err) }
	// exceeding risk limit (add one more 2000 -> 6000)
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: exp.ID, Quantity: 1}}); err == nil {
		t.Fatalf("expected risk limit error")
	}
	// reset cart by removing item
	if _, err := svc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: exp.ID}}); err != nil { t.Fatalf("remove: %v", err) }

	// max distinct items
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 1}}); err != nil { t.Fatalf("add1: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 2, Quantity: 1}}); err != nil { t.Fatalf("add2: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 3, Quantity: 1}}); err != nil { t.Fatalf("add3: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 4, Quantity: 1}}); err == nil {
		t.Fatalf("expected max distinct items error")
	}
	// total items cap (sum qty <=10)
	// clear cart
	for _, pid := range []uint{1,2,3} {
		_, _ = svc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: pid}})
	}
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 5}}); err != nil { t.Fatalf("add qty5: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 2, Quantity: 5}}); err != nil { t.Fatalf("add qty5: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 3, Quantity: 1}}); err == nil {
		t.Fatalf("expected total items cap error")
	}
	// per-line max
	// clear cart
	for _, pid := range []uint{1,2,3} { _, _ = svc.RemoveFromCart(ctx, &endpoint.HTTPRequest[*dto.RemoveFromCartRequest]{Body: &dto.RemoveFromCartRequest{UserID: 1, ProductID: pid}}) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 5}}); err != nil { t.Fatalf("add qty5: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: 1, Quantity: 1}}); err == nil {
		t.Fatalf("expected per-line limit error")
	}
	// stock availability on add
	// create low stock product
	lowResp, err := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Low", Author: "A", Description: "", Price: 10, Stock: 3}})
	if err != nil { t.Fatalf("create low: %v", err) }
	low := lowResp.Body
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: low.ID, Quantity: 2}}); err != nil { t.Fatalf("add low 2: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: low.ID, Quantity: 2}}); err == nil {
		t.Fatalf("expected stock availability error")
	}
	// low-stock per-user cap (stock<3 -> limit 1)
	veryLowResp, err := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "VeryLow", Author: "A", Description: "", Price: 10, Stock: 2}})
	if err != nil { t.Fatalf("create veryLow: %v", err) }
	veryLow := veryLowResp.Body
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 2, ProductID: veryLow.ID, Quantity: 1}}); err != nil { t.Fatalf("add 1: %v", err) }
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 2, ProductID: veryLow.ID, Quantity: 1}}); err == nil {
		t.Fatalf("expected low-stock per-user cap error")
	}
	// discontinued
	discResp, err := prodSvc.CreateProduct(ctx, &endpoint.HTTPRequest[*dto.CreateProductRequest]{Body: &dto.CreateProductRequest{Title: "Disc", Author: "A", Description: "", Price: 10, Stock: 5, Discontinued: true}})
	if err != nil { t.Fatalf("create disc: %v", err) }
	disc := discResp.Body
	if _, err := svc.AddToCart(ctx, &endpoint.HTTPRequest[*dto.AddToCartRequest]{Body: &dto.AddToCartRequest{UserID: 1, ProductID: disc.ID, Quantity: 1}}); err == nil {
		t.Fatalf("expected product unavailable error")
	}
}
