package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/endpoint"
	"ecom-book-store-sample-api/internal/storage"
)

func TestProductService_CRUD(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	svc := NewProductService(store)

	// list
	itemsResp, err := svc.ListProducts(ctx, endpoint.NewHTTPRequest(&dto.ListProductsRequest{}))
	if err != nil { t.Fatalf("list: %v", err) }
	items := itemsResp.Data
	if len(items) != 10 { t.Fatalf("expected 10 seeded products, got %d", len(items)) }

	// create
	createdResp, err := svc.CreateProduct(ctx, endpoint.NewHTTPRequest(&dto.CreateProductRequest{Title: "Test", Author: "A", Description: "D", Price: 9.99, Stock: 5}))
	if err != nil { t.Fatalf("create: %v", err) }
	created := createdResp.Data
	if created.ID == 0 { t.Fatalf("expected created ID > 0") }

	// get
	gotResp, err := svc.GetProduct(ctx, endpoint.NewHTTPRequest(&dto.GetProductRequest{ID: created.ID}))
	if err != nil { t.Fatalf("get: %v", err) }
	got := gotResp.Data
	if got.Title != "Test" { t.Fatalf("expected title Test, got %s", got.Title) }

	// update
	updResp, err := svc.UpdateProduct(ctx, endpoint.NewHTTPRequest(&dto.UpdateProductRequest{ID: created.ID, Title: "Updated", Author: "A", Description: "D2", Price: 11.99, Stock: 7}))
	if err != nil { t.Fatalf("update: %v", err) }
	upd := updResp.Data
	if upd.Title != "Updated" { t.Fatalf("expected Updated, got %s", upd.Title) }

	// delete
	if _, err := svc.DeleteProduct(ctx, endpoint.NewHTTPRequest(&dto.DeleteProductRequest{ID: created.ID})); err != nil { t.Fatalf("delete: %v", err) }
	if _, err := svc.GetProduct(ctx, endpoint.NewHTTPRequest(&dto.GetProductRequest{ID: created.ID})); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestProductService_ValidationsAndDeleteGuard(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	svc := NewProductService(store)

	// invalid price
	if _, err := svc.CreateProduct(ctx, endpoint.NewHTTPRequest(&dto.CreateProductRequest{Title: "X", Author: "A", Description: "", Price: 0, Stock: 1})); err == nil {
		t.Fatalf("expected error for price out of bounds")
	}
	// invalid title
	if _, err := svc.CreateProduct(ctx, endpoint.NewHTTPRequest(&dto.CreateProductRequest{Title: "", Author: "A", Description: "", Price: 1, Stock: 1})); err == nil {
		t.Fatalf("expected error for invalid title")
	}
	// delete guard when in cart
	cartSvc := NewCartService(store)
	createdResp, err := svc.CreateProduct(ctx, endpoint.NewHTTPRequest(&dto.CreateProductRequest{Title: "Y", Author: "A", Description: "", Price: 10, Stock: 5}))
	if err != nil { t.Fatalf("create: %v", err) }
	created := createdResp.Data
	if _, err := cartSvc.AddToCart(ctx, endpoint.NewHTTPRequest(&dto.AddToCartRequest{UserID: 1, ProductID: created.ID, Quantity: 1})); err != nil { t.Fatalf("add: %v", err) }
	if _, err := svc.DeleteProduct(ctx, endpoint.NewHTTPRequest(&dto.DeleteProductRequest{ID: created.ID})); err == nil {
		t.Fatalf("expected delete guard error")
	}
}

