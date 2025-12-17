package handler

import (
	"appointment-booking/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AvailabilityHandler struct {
	service *service.AvailabilityService
}

func NewAvailabilityHandler(service *service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{service: service}
}

func (h *AvailabilityHandler) SetAvailability(c *gin.Context) {
	var input service.SetAvailabilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	providerID := c.MustGet("userID").(uuid.UUID)
	// Ideally check if role == "provider"

	if err := h.service.SetAvailability(providerID, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Availability set"})
}

func (h *AvailabilityHandler) GetSlots(c *gin.Context) {
	providerIDStr := c.Param("providerID")
	dateStr := c.Query("date") // ?date=2025-10-30

	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	slots, err := h.service.GetAvailableSlots(providerID, dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slots": slots})
}