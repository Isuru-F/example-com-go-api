package services

import (
	"context"
	"errors"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/storage"
)

type CartService struct { store *storage.MemoryStore }

func NewCartService(store *storage.MemoryStore) *CartService { return &CartService{store: store} }

func (s *CartService) AddToCart(ctx context.Context, req *dto.AddToCartRequest) (*dto.Cart, error) {
	_ = ctx
	// Pre-validate against business rules
	p, err := s.store.GetProductByID(req.ProductID)
	if err != nil { return nil, err }
	if p.Discontinued { return nil, errors.New("product unavailable") }
	cart, err := s.store.GetCartByUser(req.UserID)
	if err != nil { return nil, err }
	// distinct items limit
	found := false
	for _, it := range cart.Items { if it.ProductID == req.ProductID { found = true; break } }
	if !found && len(cart.Items) >= MaxDistinctCartItems { return nil, errors.New("cart has too many distinct items") }
	// per-line max and stock checks
	currentQty := 0
	for _, it := range cart.Items { if it.ProductID == req.ProductID { currentQty = it.Quantity; break } }
	if currentQty+req.Quantity > MaxQuantityPerLineItem { return nil, errors.New("quantity exceeds per-item limit") }
	if p.Stock < currentQty+req.Quantity { return nil, errors.New("insufficient stock for requested quantity") }
	if p.Stock < 3 && currentQty+req.Quantity > 1 { return nil, errors.New("low-stock item limited to 1 per order") }
	// total items cap
	sumQty := 0
	for _, it := range cart.Items { sumQty += it.Quantity }
	if sumQty+req.Quantity > MaxTotalItemsInCart { return nil, errors.New("cart has too many items") }
	// risk cap (compute total using UnitPrice when present, else current price)
	total := 0.0
	for _, it := range cart.Items {
		price := it.UnitPrice
		if price == 0 { price = p.Price }
		total += float64(it.Quantity) * price
	}
	total += float64(req.Quantity) * p.Price
	if total > CartRiskLimitTotal { return nil, errors.New("cart total exceeds limit") }
	return s.store.AddToCart(req.UserID, req.ProductID, req.Quantity)
}

func (s *CartService) RemoveFromCart(ctx context.Context, req *dto.RemoveFromCartRequest) (*dto.Cart, error) {
	_ = ctx
	return s.store.RemoveFromCart(req.UserID, req.ProductID)
}

func (s *CartService) GetCart(ctx context.Context, req *dto.GetCartRequest) (*dto.Cart, error) {
	_ = ctx
	return s.store.GetCartByUser(req.UserID)
}
