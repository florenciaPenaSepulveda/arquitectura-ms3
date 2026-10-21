package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"

	// Asegúrate de que tu go.mod se llame "ms3". Si se llama diferente, ajusta estos imports.
	"ms3/db"
	"ms3/handlers"
	"ms3/models"
	"ms3/services"
)

func main() {
	// ---------------------------------------------------------
	// 1. CONEXIÓN A BASE DE DATOS
	// ---------------------------------------------------------
	// Revisa en tu archivo db/db.go cómo se llama la función.
	// Si es 'func Connect() ...', cambia Init() por Connect()
	dbInstance := db.InitDB()

	// ---------------------------------------------------------
	// 2. MIGRACIONES AUTOMÁTICAS
	// ---------------------------------------------------------
	// Esto crea la tabla 'mechanic_availabilities' automáticamente
	err := dbInstance.AutoMigrate(&models.Reservation{}, &models.MechanicAvailability{})
	if err != nil {
		log.Fatal("Error migrando la BD:", err)
	}

	// ---------------------------------------------------------
	// 3. INICIALIZACIÓN DE SERVICIOS
	// ---------------------------------------------------------
	availService := services.NewAvailabilityService(dbInstance)
	availHandler := handlers.NewAvailabilityHandler(availService)

	// Si tu ReservationHandler necesita la DB, pásasela aquí.
	// Si tu constructor es diferente, ajústalo.
	resHandler := handlers.NewReservationHandler(dbInstance)

	// ---------------------------------------------------------
	// 4. CONEXIÓN A NATS
	// ---------------------------------------------------------
	// Usamos el nombre del contenedor de NATS como host
	// Usamos el nombre del contenedor de NATS como host
	nc, err := nats.Connect("nats://nats-ms3-local:4222")
	if err != nil {
		log.Fatal("No se pudo conectar a NATS:", err)
	}
	defer nc.Close()

	// Suscribirse a mensajes de NATS (Lo que manda el test-client)
	_, err = nc.QueueSubscribe("reservation.create", "reservation-workers", func(msg *nats.Msg) {
		resHandler.HandleCreateReservation(msg)
	})
	if err != nil {
		log.Fatal("Error suscribiendo a NATS:", err)
	}

	// ---------------------------------------------------------
	// 5. SERVIDOR HTTP (GIN)
	// ---------------------------------------------------------
	r := gin.Default()

	// Configuración CORS básica
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Rutas de la Agenda
	r.POST("/availability", availHandler.CreateSchedule)
	r.GET("/availability/slots", availHandler.GetSlots)

	log.Println("Servidor MS3 corriendo en puerto :8080")
	r.Run(":8080")
}
