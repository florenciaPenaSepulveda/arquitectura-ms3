package models

import (
	"testing"
	"time"
)

func TestReservationModel(t *testing.T) {
	// Usamos tipos correctos: time.Date y uint
	fecha := time.Date(2025, 12, 12, 10, 0, 0, 0, time.UTC)

	res := Reservation{
		UserID:      "user-123",
		CarID:       "BBCL-20",
		MechanicID:  7,     // Ahora es un número (uint), no string
		Date:        fecha, // Ahora es un objeto time.Time
		StartTime:   "10:00",
		EndTime:     "11:00",
		Status:      "CONFIRMED",
		Description: "Cambio de aceite",
	}

	if res.MechanicID != 7 {
		t.Errorf("Esperaba MechanicID 7, obtuvo %d", res.MechanicID)
	}
}
