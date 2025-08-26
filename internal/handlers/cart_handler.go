package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/services"
)

type CartHandler struct { svc *services.CartService }

func NewCartHandler(svc *services.CartService) *CartHandler { return &CartHandler{svc: svc} }

type addToCartRequest struct {
	ProductID uint `json:"productId"`
	Quantity  int  `json:"quantity"`
}

type removeFromCartRequest struct {
	ProductID uint `json:"productId"`
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"}); return }
	var req addToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	cart, err := h.svc.AddToCart(userID, req.ProductID, req.Quantity)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"}); return }
	var req removeFromCartRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	cart, err := h.svc.RemoveFromCart(userID, req.ProductID)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, cart)
}
