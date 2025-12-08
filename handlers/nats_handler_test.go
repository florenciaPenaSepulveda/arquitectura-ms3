package handlers

import (
	"encoding/json"
	"ms3/models"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Función auxiliar para crear una DB de prueba rápida
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Falló conexión DB: %v", err)
	}
	db.AutoMigrate(&models.Reservation{})
	return db
}

func TestProcessCreateReservation(t *testing.T) {
	db := setupTestDB(t)

	// Simulamos el JSON que enviaría NestJS
	inputJSON := `{
		"data": {
			"user_id": "user-123",
			"car_id": "car-99",
			"date": "2025-10-10"
		},
		"id": "msg-id"
	}`

	// Probamos la lógica DIRECTAMENTE (sin NATS)
	responseBytes := ProcessCreateReservation(db, []byte(inputJSON))

	// Verificamos
	var respRes models.Reservation
	err := json.Unmarshal(responseBytes, &respRes)
	assert.NoError(t, err)
	assert.Equal(t, "PENDIENTE", respRes.Status)
	assert.Equal(t, "user-123", respRes.UserID)
}

func TestProcessGetReservations(t *testing.T) {
	db := setupTestDB(t)
	// Guardamos una reserva falsa antes de consultar
	db.Create(&models.Reservation{UserID: "test-user", Status: "CONFIRMADA"})

	responseBytes := ProcessGetReservations(db)

	var list []models.Reservation
	json.Unmarshal(responseBytes, &list)
	assert.NotEmpty(t, list)
	assert.Equal(t, "test-user", list[0].UserID)
}
