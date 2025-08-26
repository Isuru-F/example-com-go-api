package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/services"
)

type ProductHandler struct { svc *services.ProductService }

func NewProductHandler(svc *services.ProductService) *ProductHandler { return &ProductHandler{svc: svc} }

type productInput struct {
	Title        string  `json:"title"`
	Author       string  `json:"author"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Stock        int     `json:"stock"`
	Discontinued bool    `json:"discontinued"`
	IsSpecial    bool    `json:"isSpecial"`
}

var (
	prodOpsMu sync.Mutex
	prodOps   []time.Time
)

func allowProductMutation(limit int, window time.Duration) bool {
	prodOpsMu.Lock()
	defer prodOpsMu.Unlock()
	now := time.Now()
	pruned := prodOps[:0]
	for _, t := range prodOps { if now.Sub(t) <= window { pruned = append(pruned, t) } }
	if len(pruned) >= limit { prodOps = pruned; return false }
	pruned = append(pruned, now)
	prodOps = pruned
	return true
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	items, err := h.svc.ListProducts(c.Request.Context(), &dto.ListProductsRequest{})
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	if !allowProductMutation(5, time.Minute) { c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many product changes"}); return }
	var in productInput
	if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	req := &dto.CreateProductRequest{Title: in.Title, Author: in.Author, Description: in.Description, Price: in.Price, Stock: in.Stock, Discontinued: in.Discontinued, IsSpecial: in.IsSpecial}
	created, err := h.svc.CreateProduct(c.Request.Context(), req)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusCreated, created)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	p, err := h.svc.GetProduct(c.Request.Context(), &dto.GetProductRequest{ID: id})
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, p)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	if !allowProductMutation(5, time.Minute) { c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many product changes"}); return }
	id, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	var in productInput
	if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	req := &dto.UpdateProductRequest{ID: id, Title: in.Title, Author: in.Author, Description: in.Description, Price: in.Price, Stock: in.Stock, Discontinued: in.Discontinued, IsSpecial: in.IsSpecial}
	updated, err := h.svc.UpdateProduct(c.Request.Context(), req)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, updated)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	if err := h.svc.DeleteProduct(c.Request.Context(), &dto.DeleteProductRequest{ID: id}); err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
	c.Status(http.StatusNoContent)
}

func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	return uint(v), err
}
