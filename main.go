package main

import (
	"log"
	"ms3/db"       // Nota: ahora usa 'ms3' porque así se llama tu carpeta
	"ms3/handlers" // Nota: ahora usa 'ms3'
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

func main() {
	// 1. Conectar DB
	database := db.InitDB()

	// 2. Conectar NATS
	// Usamos nats.DefaultURL que suele ser "nats://127.0.0.1:4222"
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("No se pudo conectar a NATS:", err)
	}
	defer nc.Close()
	log.Println("Conectado a NATS")

	// 3. Registrar los Handlers
	handlers.CreateReservation(nc, database)
	handlers.GetReservations(nc, database)
	handlers.GetClientReservations(nc, database)
	handlers.UpdateReservation(nc, database)

	log.Println("MS3 (Reservas) corriendo y esperando mensajes...")

	// 4. Mantener el programa vivo CORRECTAMENTE
	// Creamos un canal que escuche señales del sistema (como Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// El programa se bloqueará aquí hasta recibir la señal
	<-quit

	log.Println("Apagando servidor...")
}
