package models

import (
	"gorm.io/gorm"
)

// Reservation representa la tabla en la base de datos
type Reservation struct {
	gorm.Model
	UserID      string `json:"user_id"`
	CarID       string `json:"car_id"`
	Date        string `json:"date"` // <--- CAMBIA ESTO (antes era time.Time)
	Status      string `json:"status"`
	MechanicID  string `json:"mechanic_id"`
	Description string `json:"description"`
}
