package services

import (
	"context"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/models"
	"ecom-book-store-sample-api/internal/storage"
)

type ProductService struct { store *storage.MemoryStore }

func NewProductService(store *storage.MemoryStore) *ProductService { return &ProductService{store: store} }

// Context-aware, DTO-based signatures (legacy upgrade target style)
func (s *ProductService) ListProducts(ctx context.Context, req *dto.ListProductsRequest) ([]*dto.Product, error) {
	_ = ctx // not used yet
	return s.store.GetAllProducts()
}

func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.Product, error) {
	_ = ctx
	p := &models.Product{Title: req.Title, Author: req.Author, Description: req.Description, Price: req.Price, Stock: req.Stock}
	return s.store.CreateProduct(p)
}

func (s *ProductService) GetProduct(ctx context.Context, req *dto.GetProductRequest) (*dto.Product, error) {
	_ = ctx
	return s.store.GetProductByID(req.ID)
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *dto.UpdateProductRequest) (*dto.Product, error) {
	_ = ctx
	p := &models.Product{Title: req.Title, Author: req.Author, Description: req.Description, Price: req.Price, Stock: req.Stock}
	return s.store.UpdateProduct(req.ID, p)
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *dto.DeleteProductRequest) error {
	_ = ctx
	return s.store.DeleteProduct(req.ID)
}
