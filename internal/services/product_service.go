package services

import (
	"context"
	"errors"
	"strings"

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

func validateProductInput(title, author, description string, price float64, stock int) error {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if title == "" || len(title) > 200 { return errors.New("invalid title") }
	if author == "" { return errors.New("invalid author") }
	if len(description) > 2000 { return errors.New("description too long") }
	if price < 0.01 || price > 10000 { return errors.New("price out of bounds") }
	if stock < 0 || stock > 10000 { return errors.New("invalid stock") }
	return nil
}

func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.Product, error) {
	_ = ctx
	if err := validateProductInput(req.Title, req.Author, req.Description, req.Price, req.Stock); err != nil {
		return nil, err
	}
	p := &models.Product{Title: req.Title, Author: req.Author, Description: req.Description, Price: req.Price, Stock: req.Stock, Discontinued: req.Discontinued, IsSpecial: req.IsSpecial}
	return s.store.CreateProduct(p)
}

func (s *ProductService) GetProduct(ctx context.Context, req *dto.GetProductRequest) (*dto.Product, error) {
	_ = ctx
	return s.store.GetProductByID(req.ID)
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *dto.UpdateProductRequest) (*dto.Product, error) {
	_ = ctx
	if err := validateProductInput(req.Title, req.Author, req.Description, req.Price, req.Stock); err != nil {
		return nil, err
	}
	p := &models.Product{Title: req.Title, Author: req.Author, Description: req.Description, Price: req.Price, Stock: req.Stock, Discontinued: req.Discontinued, IsSpecial: req.IsSpecial}
	return s.store.UpdateProduct(req.ID, p)
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *dto.DeleteProductRequest) error {
	_ = ctx
	if s.store.IsProductInAnyCart(req.ID) {
		return errors.New("product is present in carts")
	}
	return s.store.DeleteProduct(req.ID)
}
