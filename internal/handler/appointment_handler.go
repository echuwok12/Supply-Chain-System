package handler

import (
	"appointment-booking/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppointmentHandler struct {
	service *service.AppointmentService
}

type RescheduleInput struct {
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

func NewAppointmentHandler(service *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{service: service}
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	var input service.BookingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get User ID from JWT Context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	appointment, err := h.service.BookAppointment(userID, input)
	if err != nil {
		// Differentiate errors (logic vs server)
		if err.Error() == "time slot is not available" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

// Cancel handles POST /appointments/:id/cancel
func (h *AppointmentHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.service.CancelAppointment(id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment cancelled"})
}

// Reschedule handles POST /appointments/:id/reschedule
func (h *AppointmentHandler) Reschedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	var input RescheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.service.RescheduleAppointment(id, userID, input.StartTime, input.EndTime); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // 409 if slot taken
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment rescheduled"})
}
