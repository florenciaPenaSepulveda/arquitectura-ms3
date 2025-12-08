package models

import (
	"testing"

	"github.com/glebarez/sqlite" // Importante: Driver compatible con Windows sin CGO
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestReservationCreation(t *testing.T) {
	// 1. Configuración de la Base de Datos en Memoria
	// Usamos ":memory:" para una DB temporal y rápida.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Falló la conexión a la base de datos: %v", err)
	}

	// 2. Migración automática (Crea la tabla basada en tu struct Reservation)
	err = db.AutoMigrate(&Reservation{})
	if err != nil {
		t.Fatalf("Falló la migración: %v", err)
	}

	// 3. Preparar los datos de prueba
	// Aquí usamos los campos EXACTOS que me mostraste en reservation.go
	inputReservation := Reservation{
		UserID:      "user-123", // String, coincide con tu modelo
		CarID:       "car-ABC-99",
		Date:        "2025-12-12", // String, coincide con tu modelo
		Status:      "pending",
		MechanicID:  "mech-007",
		Description: "Revisión de frenos",
	}

	// 4. Ejecutar: Guardar en la base de datos
	result := db.Create(&inputReservation)

	// 5. Validaciones (Asserts)
	// Verificamos que no hubo error al guardar
	assert.NoError(t, result.Error, "No debería haber error al guardar la reserva")

	// Verificamos que GORM le asignó un ID automáticamente
	assert.NotZero(t, inputReservation.ID, "La reserva debería tener un ID generado")

	// 6. Validación Extra: Leer de la DB para confirmar que se guardó bien
	var foundReservation Reservation
	db.First(&foundReservation, inputReservation.ID)

	// Comparamos lo que enviamos vs lo que recuperamos
	assert.Equal(t, inputReservation.UserID, foundReservation.UserID)
	assert.Equal(t, inputReservation.Date, foundReservation.Date)
	assert.Equal(t, inputReservation.Description, foundReservation.Description)
}
