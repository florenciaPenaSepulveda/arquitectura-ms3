package main

import (
	"log"
	"ms3/db"       // Asegúrate de que tu carpeta db tenga el package db
	"ms3/handlers" // Asegúrate de que tu carpeta handlers tenga el package handlers
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

func main() {
	// 1. Inicializar la Base de Datos
	database := db.InitDB()

	// 2. Configurar la conexión a NATS (Compatible con Docker y Local)
	// Intentamos leer la variable de entorno que pusimos en el docker-compose.yml
	natsURL := os.Getenv("NATS_URL")

	// Si la variable está vacía (significa que estás corriendo local sin Docker),
	// usamos la dirección por defecto (localhost:4222)
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	// 3. Conectarse a NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("❌ Error conectando a NATS: ", err)
	}
	// Nos aseguramos de cerrar la conexión al terminar
	defer nc.Close()

	log.Println("✅ Conectado a NATS en:", natsURL)

	// 4. Registrar los Handlers (Suscribirse a los mensajes)
	// Aquí le pasamos la conexión de NATS y la base de datos a tus funciones
	handlers.CreateReservation(nc, database)
	handlers.GetReservations(nc, database)
	handlers.GetClientReservations(nc, database)
	handlers.UpdateReservation(nc, database)

	log.Println("🚀 MS3 (Reservas) corriendo y esperando mensajes...")

	// 5. Esperar señal de apagado (Ctrl + C) para cerrar limpio
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Apagando servidor...")
}
