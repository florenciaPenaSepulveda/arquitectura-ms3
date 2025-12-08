package handlers

import (
	"encoding/json"
	"log"
	"ms3/models"

	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

// CreateReservation
func CreateReservation(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.create", func(m *nats.Msg) {
		log.Println("--- DEBUG: Mensaje recibido en Go ---")
		log.Println(string(m.Data))

		// 1. DEFINIMOS EL "SOBRE" DE NESTJS
		// Creamos una estructura temporal solo para abrir el paquete
		var nestMsg struct {
			Data models.Reservation `json:"data"` // Aquí es donde vive tu reserva realmente
			ID   string             `json:"id"`
		}

		// 2. ABRIMOS EL SOBRE
		err := json.Unmarshal(m.Data, &nestMsg)
		if err != nil {
			log.Printf("--- ERROR CRÍTICO JSON: %v ---", err)
			nc.Publish(m.Reply, []byte(`{"error": "Formato incorrecto"}`))
			return
		}

		// 3. SACAMOS LA RESERVA DEL SOBRE
		newRes := nestMsg.Data

		// Ahora newRes ya tiene los datos limpios (User_id, Car_id, etc.)
		newRes.Status = "PENDIENTE"

		// 4. GUARDAR
		result := db.Create(&newRes)
		if result.Error != nil {
			log.Printf("--- ERROR BD: %v ---", result.Error)
			nc.Publish(m.Reply, []byte(`{"error": "No se pudo guardar en BD"}`))
			return
		}

		// 5. RESPONDER
		// Devolvemos la reserva creada. NestJS entenderá esto como la respuesta final.
		response, _ := json.Marshal(newRes)
		nc.Publish(m.Reply, response)
		log.Println("¡ÉXITO! Reserva guardada con ID:", newRes.ID)
	})
}

// GetReservations
func GetReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_all", func(m *nats.Msg) {
		var reservations []models.Reservation
		db.Find(&reservations)
		response, _ := json.Marshal(reservations)
		nc.Publish(m.Reply, response)
	})
}

// GetClientReservations
func GetClientReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_by_user", func(m *nats.Msg) {
		userID := string(m.Data)
		var reservations []models.Reservation

		db.Where("user_id = ?", userID).Find(&reservations)

		response, _ := json.Marshal(reservations)
		nc.Publish(m.Reply, response)
	})
} // <--- ¡ESTA LLAVE ES LA QUE TE FALTA! (Cierra la función GetClientReservations)

// UpdateReservation
// UpdateReservation (Versión corregida que abre el sobre de NestJS)
func UpdateReservation(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.update", func(m *nats.Msg) {
		log.Println("--- DEBUG: Update recibido ---")
		log.Println(string(m.Data))

		// 1. ESTRUCTURA PARA ABRIR EL SOBRE DE NESTJS
		var nestMsg struct {
			Data struct {
				ID     uint   `json:"id"`
				Status string `json:"status"`
			} `json:"data"`
			ID string `json:"id"` // ID interno de NestJS
		}

		// 2. ABRIR EL SOBRE
		if err := json.Unmarshal(m.Data, &nestMsg); err != nil {
			log.Printf("Error leyendo JSON: %v", err)
			return
		}

		updateData := nestMsg.Data // Sacamos los datos reales

		// 3. BUSCAR Y ACTUALIZAR
		var reservation models.Reservation
		if result := db.First(&reservation, updateData.ID); result.Error != nil {
			nc.Publish(m.Reply, []byte(`{"error": "Reserva no encontrada"}`))
			return
		}

		reservation.Status = updateData.Status
		db.Save(&reservation)

		log.Printf("Reserva %d actualizada a: %s", reservation.ID, reservation.Status)

		// Respondemos con la reserva actualizada
		response, _ := json.Marshal(reservation)
		nc.Publish(m.Reply, response)
	})
}
