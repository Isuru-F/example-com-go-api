package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/dto"
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

var (
	cartOpsMu sync.Mutex
	cartOps   = map[uint][]time.Time{}
)

func allowCartOp(userID uint, limit int, window time.Duration) bool {
	cartOpsMu.Lock()
	defer cartOpsMu.Unlock()
	now := time.Now()
	evts := cartOps[userID]
	// prune old
	pruned := evts[:0]
	for _, t := range evts { if now.Sub(t) <= window { pruned = append(pruned, t) } }
	if len(pruned) >= limit { cartOps[userID] = pruned; return false }
	pruned = append(pruned, now)
	cartOps[userID] = pruned
	return true
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"}); return }
	if !allowCartOp(userID, 10, time.Minute) { c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many cart updates"}); return }
	var body addToCartRequest
	if err := c.ShouldBindJSON(&body); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	req := &dto.AddToCartRequest{UserID: userID, ProductID: body.ProductID, Quantity: body.Quantity}
	cart, err := h.svc.AddToCart(c.Request.Context(), req)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"}); return }
	if !allowCartOp(userID, 10, time.Minute) { c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many cart updates"}); return }
	var body removeFromCartRequest
	if err := c.ShouldBindJSON(&body); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"}); return }
	req := &dto.RemoveFromCartRequest{UserID: userID, ProductID: body.ProductID}
	cart, err := h.svc.RemoveFromCart(c.Request.Context(), req)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, cart)
}
