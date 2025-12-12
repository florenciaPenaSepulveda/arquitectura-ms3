package handlers

import (
	"log"

	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

// 1. Definimos la estructura (Esto es lo que faltaba)
type ReservationHandler struct {
	DB *gorm.DB
}

// 2. El Constructor
func NewReservationHandler(db *gorm.DB) *ReservationHandler {
	return &ReservationHandler{DB: db}
}

// 3. Tu lógica para procesar mensajes (Ejemplo básico)
func (h *ReservationHandler) HandleCreateReservation(msg *nats.Msg) {
	log.Println("Mensaje recibido en reserva:", string(msg.Data))
	// Aquí iría tu lógica para guardar en BD...

	// Responder al NATS (si es necesario)
	msg.Respond([]byte("Reserva procesada"))
}
