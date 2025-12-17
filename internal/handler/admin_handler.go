package handler

import (
	"appointment-booking/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service *service.ReportService
}

func NewAdminHandler(service *service.ReportService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	data, err := h.service.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, data)
}
