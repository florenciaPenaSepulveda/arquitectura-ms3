package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	gorm.Model            // Esto agrega ID, CreatedAt, UpdatedAt automáticamente
	UserID      string    `json:"user_id"`
	CarID       string    `json:"car_id"`
	MechanicID  uint      `json:"mechanic_id"` // NECESARIO para la agenda
	Date        time.Time `json:"date"`
	StartTime   string    `json:"start_time"` // Ej: "10:00" (NUEVO)
	EndTime     string    `json:"end_time"`   // Ej: "11:00" (NUEVO)
	Status      string    `json:"status"`     // Ej: "CONFIRMED" (NUEVO)
	Description string    `json:"description"`
}
