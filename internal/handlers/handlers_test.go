package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/models"
	"ecom-book-store-sample-api/internal/services"
	"ecom-book-store-sample-api/internal/storage"
)

type productResp models.Product

type orderResp models.Order

type cartResp models.Cart

func setupRouter() (*gin.Engine, *storage.MemoryStore) {
	gin.SetMode(gin.TestMode)
	store := storage.NewMemoryStore()
	storage.Seed(store)

	productSvc := services.NewProductService(store)
	cartSvc := services.NewCartService(store)
	orderSvc := services.NewOrderService(store)

	r := gin.New()
	api := r.Group("/api/v1")
	{
		ph := NewProductHandler(productSvc)
		api.GET("/products", ph.ListProducts)
		api.POST("/products", ph.CreateProduct)
		api.GET("/products/:id", ph.GetProduct)
		api.PUT("/products/:id", ph.UpdateProduct)
		api.DELETE("/products/:id", ph.DeleteProduct)

		ch := NewCartHandler(cartSvc)
		api.POST("/cart/user/:id/items", ch.AddToCart)
		api.DELETE("/cart/user/:id/items", ch.RemoveFromCart)

		oh := NewOrderHandler(orderSvc)
		api.POST("/orders/user/:id", oh.PlaceOrder)
	}
	return r, store
}

func do(r *gin.Engine, method, path string, body string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" { reader = bytes.NewReader(nil) } else { reader = bytes.NewReader([]byte(body)) }
	req := httptest.NewRequest(method, path, reader)
	if body != "" { req.Header.Set("Content-Type", "application/json") }
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestListProducts(t *testing.T) {
	r, _ := setupRouter()
	rec := do(r, http.MethodGet, "/api/v1/products", "")
	if rec.Code != http.StatusOK { t.Fatalf("expected 200, got %d", rec.Code) }
	var out []productResp
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil { t.Fatalf("json: %v", err) }
	if len(out) != 10 { t.Fatalf("expected 10 products, got %d", len(out)) }
}

func TestProductCRUD(t *testing.T) {
	r, _ := setupRouter()
	// create
	rec := do(r, http.MethodPost, "/api/v1/products", `{"title":"Test","author":"A","description":"D","price":9.99,"stock":5}`)
	if rec.Code != http.StatusCreated { t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String()) }
	var created productResp
	json.Unmarshal(rec.Body.Bytes(), &created)
	// get
	rec = do(r, http.MethodGet, "/api/v1/products/"+itoa(created.ID), "")
	if rec.Code != http.StatusOK { t.Fatalf("get: expected 200, got %d", rec.Code) }
	// update
	rec = do(r, http.MethodPut, "/api/v1/products/"+itoa(created.ID), `{"title":"Updated","author":"A","description":"D2","price":11.99,"stock":7}`)
	if rec.Code != http.StatusOK { t.Fatalf("update: expected 200, got %d", rec.Code) }
	// delete
	rec = do(r, http.MethodDelete, "/api/v1/products/"+itoa(created.ID), "")
	if rec.Code != http.StatusNoContent { t.Fatalf("delete: expected 204, got %d", rec.Code) }
	// get after delete
	rec = do(r, http.MethodGet, "/api/v1/products/"+itoa(created.ID), "")
	if rec.Code != http.StatusNotFound { t.Fatalf("get after delete: expected 404, got %d", rec.Code) }
}

func TestCartAndOrderFlow(t *testing.T) {
	r, _ := setupRouter()
	// add to cart user 1 product 1 qty 2
	rec := do(r, http.MethodPost, "/api/v1/cart/user/1/items", `{"productId":1,"quantity":2}`)
	if rec.Code != http.StatusOK { t.Fatalf("add: expected 200, got %d: %s", rec.Code, rec.Body.String()) }
	// place order
	rec = do(r, http.MethodPost, "/api/v1/orders/user/1", "")
	if rec.Code != http.StatusCreated { t.Fatalf("order: expected 201, got %d: %s", rec.Code, rec.Body.String()) }
	var ord orderResp
	if err := json.Unmarshal(rec.Body.Bytes(), &ord); err != nil { t.Fatalf("json: %v", err) }
	if len(ord.Items) != 1 { t.Fatalf("expected 1 order item, got %d", len(ord.Items)) }
	if ord.Items[0].Quantity != 2 { t.Fatalf("expected qty 2, got %d", ord.Items[0].Quantity) }
	if ord.Total <= 0 { t.Fatalf("expected positive total, got %f", ord.Total) }
}

// helpers
func itoa(u uint) string { return fmt.Sprintf("%d", u) }
