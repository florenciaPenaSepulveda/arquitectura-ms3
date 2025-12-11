package db

import (
	"log"
	"ms3/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	// 1. Configuración de conexión
	// HOST: "postgres-db" (coincide con el nombre en docker-compose)
	// PORT: "5432" (puerto interno original de Postgres)
	dsn := "host=postgres-db user=user_ms3 password=password_ms3 dbname=reservations_db port=5432 sslmode=disable"

	var database *gorm.DB
	var err error

	// 2. Reintentos (Wait-for-it casero)
	// A veces la DB tarda unos segundos en arrancar, así que intentamos 5 veces.
	for i := 0; i < 5; i++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Println("⏳ Esperando a que la Base de Datos esté lista...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("❌ Fallo definitivo al conectar a la base de datos:", err)
	}

	// 3. Migración
	database.AutoMigrate(&models.Reservation{})

	log.Println("✅ Conectado a Postgres Dockerizado (Host: postgres-db)")
	return database
}
