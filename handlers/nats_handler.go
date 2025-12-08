package handlers

import (
	"encoding/json"
	"log"
	"ms3/models"

	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

// --- CONEXIONES (Difíciles de testear, se quedan aquí) ---

func CreateReservation(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.create", func(m *nats.Msg) {
		response := ProcessCreateReservation(db, m.Data) // Llamamos a la lógica separada
		nc.Publish(m.Reply, response)
	})
}

func GetReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_all", func(m *nats.Msg) {
		response := ProcessGetReservations(db)
		nc.Publish(m.Reply, response)
	})
}

func GetClientReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_by_user", func(m *nats.Msg) {
		response := ProcessGetClientReservations(db, string(m.Data))
		nc.Publish(m.Reply, response)
	})
}

func UpdateReservation(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.update", func(m *nats.Msg) {
		response := ProcessUpdateReservation(db, m.Data)
		nc.Publish(m.Reply, response)
	})
}

// --- LÓGICA PURA (Fáciles de testear, estas nos darán el 80%) ---

func ProcessCreateReservation(db *gorm.DB, data []byte) []byte {
	log.Println("--- DEBUG: Mensaje recibido en Go ---")

	// Estructura para abrir el mensaje de NestJS
	var nestMsg struct {
		Data models.Reservation `json:"data"`
		ID   string             `json:"id"`
	}

	if err := json.Unmarshal(data, &nestMsg); err != nil {
		log.Printf("--- ERROR CRÍTICO JSON: %v ---", err)
		return []byte(`{"error": "Formato incorrecto"}`)
	}

	newRes := nestMsg.Data
	newRes.Status = "PENDIENTE"

	if result := db.Create(&newRes); result.Error != nil {
		log.Printf("--- ERROR BD: %v ---", result.Error)
		return []byte(`{"error": "No se pudo guardar en BD"}`)
	}

	response, _ := json.Marshal(newRes)
	return response
}

func ProcessGetReservations(db *gorm.DB) []byte {
	var reservations []models.Reservation
	db.Find(&reservations)
	response, _ := json.Marshal(reservations)
	return response
}

func ProcessGetClientReservations(db *gorm.DB, userID string) []byte {
	var reservations []models.Reservation
	db.Where("user_id = ?", userID).Find(&reservations)
	response, _ := json.Marshal(reservations)
	return response
}

func ProcessUpdateReservation(db *gorm.DB, data []byte) []byte {
	var nestMsg struct {
		Data struct {
			ID     uint   `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
		ID string `json:"id"`
	}

	if err := json.Unmarshal(data, &nestMsg); err != nil {
		return []byte(`{"error": "Formato JSON incorrecto"}`)
	}

	updateData := nestMsg.Data
	var reservation models.Reservation

	if result := db.First(&reservation, updateData.ID); result.Error != nil {
		return []byte(`{"error": "Reserva no encontrada"}`)
	}

	reservation.Status = updateData.Status
	db.Save(&reservation)

	response, _ := json.Marshal(reservation)
	return response
}
