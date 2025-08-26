package services

import (
	"context"
	"testing"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

func TestProductService_CRUD(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	storage.Seed(store)
	svc := NewProductService(store)

	// list
	items, err := svc.ListProducts(ctx, &dto.ListProductsRequest{})
	if err != nil { t.Fatalf("list: %v", err) }
	if len(items) != 10 { t.Fatalf("expected 10 seeded products, got %d", len(items)) }

	// create
	created, err := svc.CreateProduct(ctx, &dto.CreateProductRequest{Title: "Test", Author: "A", Description: "D", Price: 9.99, Stock: 5})
	if err != nil { t.Fatalf("create: %v", err) }
	if created.ID == 0 { t.Fatalf("expected created ID > 0") }

	// get
	got, err := svc.GetProduct(ctx, &dto.GetProductRequest{ID: created.ID})
	if err != nil { t.Fatalf("get: %v", err) }
	if got.Title != "Test" { t.Fatalf("expected title Test, got %s", got.Title) }

	// update
	upd, err := svc.UpdateProduct(ctx, &dto.UpdateProductRequest{ID: created.ID, Title: "Updated", Author: "A", Description: "D2", Price: 11.99, Stock: 7})
	if err != nil { t.Fatalf("update: %v", err) }
	if upd.Title != "Updated" { t.Fatalf("expected Updated, got %s", upd.Title) }

	// delete
	if err := svc.DeleteProduct(ctx, &dto.DeleteProductRequest{ID: created.ID}); err != nil { t.Fatalf("delete: %v", err) }
	if _, err := svc.GetProduct(ctx, &dto.GetProductRequest{ID: created.ID}); err == nil {
		t.Fatalf("expected error after delete")
	}
}
