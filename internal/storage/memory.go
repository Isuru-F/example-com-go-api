package storage

import (
	"errors"
	"sort"
	"sync"
	"time"

	"ecom-book-store-sample-api/internal/models"
)

type MemoryStore struct {
	mu sync.RWMutex

	users       map[uint]*models.User
	products    map[uint]*models.Product
	carts       map[uint]*models.Cart     // keyed by userID
	orders      map[uint]*models.Order
	nextUserID   uint
	nextProductID uint
	nextCartID    uint
	nextOrderID   uint
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[uint]*models.User),
		products:     make(map[uint]*models.Product),
		carts:        make(map[uint]*models.Cart),
		orders:       make(map[uint]*models.Order),
		nextUserID:    1,
		nextProductID: 1,
		nextCartID:    1,
		nextOrderID:   1,
	}
}

// Users
func (m *MemoryStore) CreateUser(u *models.User) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u.ID = m.nextUserID
	m.nextUserID++
	m.users[u.ID] = &models.User{ID: u.ID, Email: u.Email, Name: u.Name}
	return m.users[u.ID], nil
}

func (m *MemoryStore) GetUserByID(id uint) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return cloneUser(u), nil
}

// Products
func (m *MemoryStore) GetAllProducts() ([]*models.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*models.Product, 0, len(m.products))
	for _, p := range m.products {
		res = append(res, cloneProduct(p))
	}
	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })
	return res, nil
}

func (m *MemoryStore) GetProductByID(id uint) (*models.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return cloneProduct(p), nil
}

func (m *MemoryStore) CreateProduct(p *models.Product) (*models.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.ID = m.nextProductID
	m.nextProductID++
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	m.products[p.ID] = &models.Product{
		ID:          p.ID,
		Title:       p.Title,
		Author:      p.Author,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	return cloneProduct(m.products[p.ID]), nil
}

func (m *MemoryStore) UpdateProduct(id uint, update *models.Product) (*models.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	existing.Title = update.Title
	existing.Author = update.Author
	existing.Description = update.Description
	existing.Price = update.Price
	existing.Stock = update.Stock
	existing.UpdatedAt = time.Now()
	return cloneProduct(existing), nil
}

func (m *MemoryStore) DeleteProduct(id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.products[id]; !ok {
		return errors.New("product not found")
	}
	delete(m.products, id)
	return nil
}

// Carts
func (m *MemoryStore) getOrCreateCart(userID uint) *models.Cart {
	c, ok := m.carts[userID]
	if !ok {
		c = &models.Cart{ID: m.nextCartID, UserID: userID, Items: []models.CartItem{}}
		m.nextCartID++
		m.carts[userID] = c
	}
	return c
}

func (m *MemoryStore) AddToCart(userID, productID uint, quantity int) (*models.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[userID]; !ok {
		return nil, errors.New("user not found")
	}
	p, ok := m.products[productID]
	if !ok {
		return nil, errors.New("product not found")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	c := m.getOrCreateCart(userID)
	// add or increment
	found := false
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			c.Items[i].Quantity += quantity
			found = true
			break
		}
	}
	if !found {
		c.Items = append(c.Items, models.CartItem{ProductID: productID, Quantity: quantity})
	}
	// Optional soft check: cap at available stock but do not fail
	if cItemQty := cartQtyForProduct(c, productID); cItemQty > p.Stock {
		// keep as-is; strict validation happens at order time
	}
	return cloneCart(c), nil
}

func (m *MemoryStore) RemoveFromCart(userID, productID uint) (*models.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.carts[userID]
	if !ok {
		return nil, errors.New("cart not found")
	}
	items := c.Items[:0]
	for _, it := range c.Items {
		if it.ProductID != productID {
			items = append(items, it)
		}
	}
	c.Items = items
	return cloneCart(c), nil
}

func (m *MemoryStore) GetCartByUser(userID uint) (*models.Cart, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.carts[userID]
	if !ok {
		return &models.Cart{ID: 0, UserID: userID, Items: []models.CartItem{}}, nil
	}
	return cloneCart(c), nil
}

func (m *MemoryStore) clearCart(userID uint) {
	delete(m.carts, userID)
}

// Orders
func (m *MemoryStore) CreateOrder(o *models.Order) (*models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o.ID = m.nextOrderID
	m.nextOrderID++
	o.CreatedAt = time.Now()
	m.orders[o.ID] = &models.Order{
		ID:        o.ID,
		UserID:    o.UserID,
		Items:     append([]models.OrderItem(nil), o.Items...),
		Total:     o.Total,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
	}
	return cloneOrder(m.orders[o.ID]), nil
}

// Business helpers used by services
func (m *MemoryStore) ReserveStockForOrder(userID uint) (*models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.carts[userID]
	if !ok || len(c.Items) == 0 {
		return nil, errors.New("cart is empty")
	}
	// Validate and reserve (decrement stock)
	items := make([]models.OrderItem, 0, len(c.Items))
	var total float64
	for _, it := range c.Items {
		p, ok := m.products[it.ProductID]
		if !ok {
			return nil, errors.New("product not found in cart")
		}
		if it.Quantity <= 0 {
			return nil, errors.New("invalid cart item quantity")
		}
		if p.Stock < it.Quantity {
			return nil, errors.New("insufficient stock for product")
		}
		p.Stock -= it.Quantity
		sub := float64(it.Quantity) * p.Price
		total += sub
		items = append(items, models.OrderItem{ProductID: p.ID, Quantity: it.Quantity, UnitPrice: p.Price, Subtotal: sub})
	}
	order := &models.Order{
		UserID: userID,
		Items:  items,
		Total:  total,
		Status: "PLACED",
	}
	// clear cart after reserving stock; the transaction is in-memory and mutex-guarded
	delete(m.carts, userID)
	return order, nil
}

// clones to avoid exposing internal pointers/state
func cloneUser(u *models.User) *models.User { v := *u; return &v }
func cloneProduct(p *models.Product) *models.Product { v := *p; return &v }
func cloneCart(c *models.Cart) *models.Cart { v := *c; v.Items = append([]models.CartItem(nil), c.Items...); return &v }
func cloneOrder(o *models.Order) *models.Order { v := *o; v.Items = append([]models.OrderItem(nil), o.Items...); return &v }

func cartQtyForProduct(c *models.Cart, productID uint) int {
	for _, it := range c.Items {
		if it.ProductID == productID { return it.Quantity }
	}
	return 0
}
