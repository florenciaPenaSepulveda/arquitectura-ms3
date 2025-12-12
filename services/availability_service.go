package services

import (
	"fmt"
	"ms3/models" // <--- REVISA QUE ESTO COINCIDA CON TU go.mod
	"time"

	"gorm.io/gorm"
)

type AvailabilityService struct {
	DB *gorm.DB
}

func NewAvailabilityService(db *gorm.DB) *AvailabilityService {
	return &AvailabilityService{DB: db}
}

// 1. Guardar horario base
func (s *AvailabilityService) SetSchedule(schedule *models.MechanicAvailability) error {
	return s.DB.Create(schedule).Error
}

// 2. Calcular Slots Disponibles
func (s *AvailabilityService) GetAvailableSlots(mechanicID uint, dateStr string) ([]models.TimeSlot, error) {
	// Parsear fecha
	dateParsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("fecha inválida")
	}
	dayOfWeek := int(dateParsed.Weekday())

	// Buscar horario base del mecánico
	var schedule models.MechanicAvailability
	// Buscamos si trabaja ese día
	err = s.DB.Where("mechanic_id = ? AND day_of_week = ? AND is_active = ?", mechanicID, dayOfWeek, true).First(&schedule).Error
	if err != nil {
		return []models.TimeSlot{}, nil // No trabaja o no configurado
	}

	// Buscar reservas YA hechas (Usamos tu modelo Reservation existente)
	var reservations []models.Reservation
	// Asegúrate de que tu modelo Reservation tenga los campos Date, StartTime, EndTime, Status
	s.DB.Where("mechanic_id = ? AND date = ? AND status = ?", mechanicID, dateParsed, "CONFIRMED").Find(&reservations)

	return generateSlots(schedule, reservations, dateStr), nil
}

// Lógica interna para cruzar horarios
func generateSlots(schedule models.MechanicAvailability, reservations []models.Reservation, dateStr string) []models.TimeSlot {
	var slots []models.TimeSlot
	serviceDuration := 60 * time.Minute // Bloques de 1 hora

	start, _ := time.Parse("2006-01-02 15:04", dateStr+" "+schedule.StartTime)
	end, _ := time.Parse("2006-01-02 15:04", dateStr+" "+schedule.EndTime)

	for current := start; current.Add(serviceDuration).Before(end) || current.Add(serviceDuration).Equal(end); current = current.Add(serviceDuration) {
		slotEnd := current.Add(serviceDuration)
		isBusy := false

		for _, res := range reservations {
			// Adaptar esto según el formato de tus campos en Reservation
			// Asumo que en Reservation guardas StartTime como string "HH:MM"
			resStart, _ := time.Parse("2006-01-02 15:04", dateStr+" "+res.StartTime)
			resEnd, _ := time.Parse("2006-01-02 15:04", dateStr+" "+res.EndTime)

			if current.Before(resEnd) && slotEnd.After(resStart) {
				isBusy = true
				break
			}
		}

		slots = append(slots, models.TimeSlot{
			Start:     current.Format("15:04"),
			End:       slotEnd.Format("15:04"),
			Available: !isBusy,
		})
	}
	return slots
}
