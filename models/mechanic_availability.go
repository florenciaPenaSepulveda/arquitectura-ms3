package models

import (
	"time"
)

// MechanicAvailability: Define el horario base (ej: Lunes de 9 a 18)
type MechanicAvailability struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	MechanicID uint      `json:"mechanic_id"`
	DayOfWeek  int       `json:"day_of_week"` // 0=Domingo, 1=Lunes...
	StartTime  string    `json:"start_time"`  // "09:00"
	EndTime    string    `json:"end_time"`    // "18:00"
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`
}

// TimeSlot: Estructura auxiliar para responder al Front (No se guarda en BD)
type TimeSlot struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Available bool   `json:"available"`
}
