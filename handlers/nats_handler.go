package handlers

import (
	"encoding/json"
	"log"
	"ms3/models"
	"strconv" // <--- AGREGADO: Necesario para convertir texto a número
	"strings"

	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

type NestEnvelope struct {
	Pattern string          `json:"pattern"`
	Data    json.RawMessage `json:"data"`
	ID      string          `json:"id"`
}

// unwrapNestEnvelope devuelve (requestId, innerData).
// Si no viene en formato Nest, devuelve ("", payloadOriginal).
func unwrapNestEnvelope(b []byte) (string, []byte) {
	var env NestEnvelope
	if err := json.Unmarshal(b, &env); err == nil && env.Pattern != "" && env.Data != nil {
		return env.ID, env.Data
	}
	return "", b
}

// respondNest responde en formato que Nest espera:
// { id: "...", response: <json>, isDisposed: true }
func respondNest(m *nats.Msg, reqID string, response []byte) error {
	// si no me pasaron reqID, intento sacarlo del envelope original
	if reqID == "" {
		var env NestEnvelope
		if err := json.Unmarshal(m.Data, &env); err == nil && env.ID != "" {
			reqID = env.ID
		}
	}

	// si de verdad no hay reqID, respondo “a pelo”
	if reqID == "" {
		return m.Respond(response)
	}

	out := struct {
		ID         string          `json:"id"`
		Response   json.RawMessage `json:"response"`
		IsDisposed bool            `json:"isDisposed"`
	}{
		ID:         reqID,
		Response:   json.RawMessage(response),
		IsDisposed: true,
	}

	b, _ := json.Marshal(out)
	return m.Respond(b)
}

func decodeString(b []byte) string {
	// intenta JSON string primero: "abc"
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		return s
	}
	// fallback
	return strings.Trim(string(b), "\"")
}

// -------------------- SUBSCRIBERS --------------------

func CreateReservation(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.create", func(m *nats.Msg) {
		reqID, inner := unwrapNestEnvelope(m.Data)
		response := ProcessCreateReservation(db, inner)
		_ = respondNest(m, reqID, response)
	})
}

func GetReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_all", func(m *nats.Msg) {
		reqID, _ := unwrapNestEnvelope(m.Data) // inner no importa (es {})
		response := ProcessGetReservations(db)
		_ = respondNest(m, reqID, response)
	})
}

func GetClientReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_by_user", func(m *nats.Msg) {
		reqID, inner := unwrapNestEnvelope(m.Data)
		userID := decodeString(inner)
		response := ProcessGetClientReservations(db, userID)
		_ = respondNest(m, reqID, response)
	})
}

func UpdateReservation(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.update", func(m *nats.Msg) {
		reqID, inner := unwrapNestEnvelope(m.Data)
		response := ProcessUpdateReservation(db, inner)
		_ = respondNest(m, reqID, response)
	})
}

func GetPendingReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_pending", func(m *nats.Msg) {
		reqID, _ := unwrapNestEnvelope(m.Data)
		response := ProcessGetPendingReservations(db)
		_ = respondNest(m, reqID, response)
	})
}

func GetMechanicReservations(nc *nats.Conn, db *gorm.DB) {
	nc.Subscribe("reservation.get_by_mechanic", func(m *nats.Msg) {
		reqID, inner := unwrapNestEnvelope(m.Data)
		mechanicID := decodeString(inner)

		var reservations []models.Reservation
		db.Where("mechanic_id = ?", mechanicID).Find(&reservations)

		response, _ := json.Marshal(reservations)
		_ = respondNest(m, reqID, response)
	})
}

// -------------------- LÓGICA “PURA” --------------------

func ProcessCreateReservation(db *gorm.DB, data []byte) []byte {
	log.Println("--- DEBUG: Mensaje recibido en Go ---")

	// 1) formato directo
	var direct models.Reservation
	if err := json.Unmarshal(data, &direct); err == nil && direct.UserID != "" {
		direct.Status = "PENDIENTE"
		if result := db.Create(&direct); result.Error != nil {
			return []byte(`{"error":"No se pudo guardar en BD"}`)
		}
		response, _ := json.Marshal(direct)
		return response
	}

	// 2) formato envuelto (data: {...})
	var wrapped struct {
		Data models.Reservation `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		log.Printf("--- ERROR CRÍTICO JSON: %v ---", err)
		return []byte(`{"error":"Formato incorrecto"}`)
	}

	newRes := wrapped.Data
	newRes.Status = "PENDIENTE"

	if result := db.Create(&newRes); result.Error != nil {
		return []byte(`{"error":"No se pudo guardar en BD"}`)
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

func ProcessGetPendingReservations(db *gorm.DB) []byte {
	var reservations []models.Reservation
	db.Where("status = ?", "PENDIENTE").Find(&reservations)
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
	log.Printf("[UPDATE] inner payload: %s", string(data))

	// Forma esperada: {"data": {"id":1,"status":"...","mechanic_id":"..."}}
	var wrapped struct {
		Data struct {
			ID         uint   `json:"id"`
			Status     string `json:"status"`
			MechanicID string `json:"mechanic_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &wrapped); err != nil || wrapped.Data.ID == 0 {
		log.Printf("[UPDATE] ERROR json: %v", err)
		return []byte(`{"error":"Formato JSON incorrecto"}`)
	}

	upd := wrapped.Data

	var r models.Reservation
	if err := db.First(&r, upd.ID).Error; err != nil {
		return []byte(`{"error":"Reserva no encontrada"}`)
	}

	r.Status = upd.Status

	// --- CORRECCIÓN AQUÍ ---
	if upd.MechanicID != "" {
		// Convertimos el string a número (base 10, 64 bits)
		idConvertido, err := strconv.ParseUint(upd.MechanicID, 10, 64)
		if err != nil {
			log.Printf("Error convirtiendo MechanicID: %v", err)
			// Podrías retornar error, o simplemente ignorar el campo si viene mal
		} else {
			// Asignamos el valor convertido a uint
			r.MechanicID = uint(idConvertido)
		}
	}
	// -----------------------

	if err := db.Save(&r).Error; err != nil {
		return []byte(`{"error":"No se pudo actualizar"}`)
	}

	response, _ := json.Marshal(r)
	return response
}
