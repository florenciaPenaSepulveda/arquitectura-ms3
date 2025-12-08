package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Estructura igual a la de tu modelo para enviar datos
type Reservation struct {
	UserID      string    `json:"user_id"`
	CarID       string    `json:"car_id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

func main() {
	// 1. Conectar a NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// 2. Crear datos de prueba (Simulando lo que enviaría el Frontend)
	nuevaReserva := Reservation{
		UserID:      "usuario_123",
		CarID:       "Patente-BBCL20",
		Date:        time.Now().Add(24 * time.Hour), // Mañana
		Description: "El auto hace un ruido extraño al frenar",
	}

	jsonBytes, _ := json.Marshal(nuevaReserva)

	// 3. Enviar mensaje al canal "reservation.create" y esperar respuesta (Request)
	// Esperamos máximo 2 segundos por una respuesta
	msg, err := nc.Request("reservation.create", jsonBytes, 2*time.Second)

	if err != nil {
		log.Fatal("Error, el MS3 no respondió: ", err)
	}

	fmt.Printf("¡Éxito! Respuesta del MS3: %s\n", string(msg.Data))
}
