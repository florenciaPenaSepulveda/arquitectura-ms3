package db

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	// Asegúrate de que el host sea el correcto (ms3-postgres)
	dsn := "host=ms3-postgres user=postgres password=secret dbname=reservations_db port=5432 sslmode=disable"

	var db *gorm.DB
	var err error

	// Intentamos conectar hasta 10 veces (20 segundos en total)
	for i := 1; i <= 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("¡Conexión a Base de Datos exitosa!")
			return db
		}

		log.Printf("Intento %d/10 fallo. La BD aún se está iniciando... Esperando 2s", i)
		time.Sleep(2 * time.Second)
	}

	// Si después de 10 intentos falla, ahí sí morimos
	log.Fatal("Fallo definitivo al conectar a la BD:", err)
	return nil
}
