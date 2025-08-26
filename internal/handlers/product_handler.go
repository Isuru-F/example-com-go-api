package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/services"
)

type ProductHandler struct { svc *services.ProductService }

func NewProductHandler(svc *services.ProductService) *ProductHandler { return &ProductHandler{svc: svc} }

type productInput struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	items, err := h.svc.ListProducts(c.Request.Context(), &dto.ListProductsRequest{})
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var in productInput
	if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	req := &dto.CreateProductRequest{Title: in.Title, Author: in.Author, Description: in.Description, Price: in.Price, Stock: in.Stock}
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
	id, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
	var in productInput
	if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	req := &dto.UpdateProductRequest{ID: id, Title: in.Title, Author: in.Author, Description: in.Description, Price: in.Price, Stock: in.Stock}
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
