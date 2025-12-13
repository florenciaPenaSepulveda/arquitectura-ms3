package main

import (
	"log"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"

	// Asegúrate de que tu go.mod se llame "ms3". Si se llama diferente, ajusta estos imports.
	"ms3/db"
	"ms3/handlers"
	"ms3/models"
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
	//availService := services.NewAvailabilityService(dbInstance)
	//availHandler := handlers.NewAvailabilityHandler(availService)

	// ---------------------------------------------------------
	// 4. CONEXIÓN A NATS
	// ---------------------------------------------------------
	// Usamos el nombre del contenedor de NATS como host
	// Usamos el nombre del contenedor de NATS como host
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222" // service name
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("No se pudo conectar a NATS:", err)
	}
	defer nc.Close()

	handlers.CreateReservation(nc, dbInstance)
	handlers.GetReservations(nc, dbInstance)
	handlers.GetClientReservations(nc, dbInstance)
	handlers.UpdateReservation(nc, dbInstance)
	handlers.GetPendingReservations(nc, dbInstance)
	handlers.GetMechanicReservations(nc, dbInstance)

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
	//r.POST("/availability", availHandler.CreateSchedule)
	//r.GET("/availability/slots", availHandler.GetSlots)

	log.Println("Servidor MS3 corriendo en puerto :8080")
	r.Run(":8080")
}
