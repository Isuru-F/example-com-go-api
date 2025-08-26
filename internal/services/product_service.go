package services

import (
	"ecom-book-store-sample-api/internal/models"
	"ecom-book-store-sample-api/internal/storage"
)

type ProductService struct {
	store *storage.MemoryStore
}

func NewProductService(store *storage.MemoryStore) *ProductService {
	return &ProductService{store: store}
}

func (s *ProductService) ListProducts() ([]*models.Product, error) {
	return s.store.GetAllProducts()
}

func (s *ProductService) CreateProduct(product *models.Product) (*models.Product, error) {
	return s.store.CreateProduct(product)
}

func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	return s.store.GetProductByID(id)
}

func (s *ProductService) UpdateProduct(id uint, product *models.Product) (*models.Product, error) {
	return s.store.UpdateProduct(id, product)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.store.DeleteProduct(id)
}
