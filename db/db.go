package db

import (
	"log"
	"ms3/models" // Importamos tu modelo

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	// String de conexión (DSN)
	dsn := "host=localhost user=user_ms3 password=password_ms3 dbname=reservations_db port=5433 sslmode=disable"

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Fallo al conectar a la base de datos:", err)
	}

	// Migración automática: Crea la tabla si no existe
	database.AutoMigrate(&models.Reservation{})

	log.Println("Base de datos conectada y migrada")
	return database
}
