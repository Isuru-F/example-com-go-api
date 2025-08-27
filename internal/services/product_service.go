package services

import (
	"context"
	"errors"
	"strings"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/endpoint"
	"ecom-book-store-sample-api/internal/models"
	"ecom-book-store-sample-api/internal/storage"
)

type ProductService struct { store *storage.MemoryStore }

func NewProductService(store *storage.MemoryStore) *ProductService { return &ProductService{store: store} }

// Updated endpoint-style signatures
func (s *ProductService) ListProducts(ctx context.Context, req *endpoint.HTTPRequest[*dto.ListProductsRequest]) (*endpoint.HTTPResponse[[]*dto.Product], error) {
	_ = ctx // not used yet
	items, err := s.store.GetAllProducts()
	if err != nil { return nil, err }
	return &endpoint.HTTPResponse[[]*dto.Product]{Body: items}, nil
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

func (s *ProductService) CreateProduct(ctx context.Context, req *endpoint.HTTPRequest[*dto.CreateProductRequest]) (*endpoint.HTTPResponse[*dto.Product], error) {
	_ = ctx
	if err := validateProductInput(req.Body.Title, req.Body.Author, req.Body.Description, req.Body.Price, req.Body.Stock); err != nil {
		return nil, err
	}
	p := &models.Product{Title: req.Body.Title, Author: req.Body.Author, Description: req.Body.Description, Price: req.Body.Price, Stock: req.Body.Stock, Discontinued: req.Body.Discontinued, IsSpecial: req.Body.IsSpecial}
	out, err := s.store.CreateProduct(p)
	if err != nil { return nil, err }
	return &endpoint.HTTPResponse[*dto.Product]{Body: out}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *endpoint.HTTPRequest[*dto.GetProductRequest]) (*endpoint.HTTPResponse[*dto.Product], error) {
	_ = ctx
	p, err := s.store.GetProductByID(req.Body.ID)
	if err != nil { return nil, err }
	return &endpoint.HTTPResponse[*dto.Product]{Body: p}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *endpoint.HTTPRequest[*dto.UpdateProductRequest]) (*endpoint.HTTPResponse[*dto.Product], error) {
	_ = ctx
	if err := validateProductInput(req.Body.Title, req.Body.Author, req.Body.Description, req.Body.Price, req.Body.Stock); err != nil {
		return nil, err
	}
	p := &models.Product{Title: req.Body.Title, Author: req.Body.Author, Description: req.Body.Description, Price: req.Body.Price, Stock: req.Body.Stock, Discontinued: req.Body.Discontinued, IsSpecial: req.Body.IsSpecial}
	out, err := s.store.UpdateProduct(req.Body.ID, p)
	if err != nil { return nil, err }
	return &endpoint.HTTPResponse[*dto.Product]{Body: out}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *endpoint.HTTPRequest[*dto.DeleteProductRequest]) (*endpoint.HTTPResponse[struct{}], error) {
	_ = ctx
	if s.store.IsProductInAnyCart(req.Body.ID) {
		return nil, errors.New("product is present in carts")
	}
	if err := s.store.DeleteProduct(req.Body.ID); err != nil { return nil, err }
	return &endpoint.HTTPResponse[struct{}]{Body: struct{}{}}, nil
}
