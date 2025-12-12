package handlers

import (
	"ms3/models"   // <--- REVISA TU go.mod
	"ms3/services" // <--- REVISA TU go.mod
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin" // Asumo que usas Gin por lo común que es
)

type AvailabilityHandler struct {
	Service *services.AvailabilityService
}

func NewAvailabilityHandler(s *services.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{Service: s}
}

func (h *AvailabilityHandler) CreateSchedule(c *gin.Context) {
	var schedule models.MechanicAvailability
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.SetSchedule(&schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Horario guardado"})
}

func (h *AvailabilityHandler) GetSlots(c *gin.Context) {
	mechanicID, _ := strconv.Atoi(c.Query("mechanicId"))
	date := c.Query("date")
	slots, _ := h.Service.GetAvailableSlots(uint(mechanicID), date)
	c.JSON(http.StatusOK, slots)
}
