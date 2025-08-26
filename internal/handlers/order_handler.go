package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecom-book-store-sample-api/internal/dto"
	"ecom-book-store-sample-api/internal/services"
)

type OrderHandler struct { svc *services.OrderService }

func NewOrderHandler(svc *services.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"}); return }
	order, err := h.svc.PlaceOrder(c.Request.Context(), &dto.PlaceOrderRequest{UserID: userID})
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusCreated, order)
}
