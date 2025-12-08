package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test 1: Caso Exitoso (POST Correcto)
func TestCreateReservationHandler(t *testing.T) {
	// 1. ARRANGE
	req, err := http.NewRequest("POST", "/reservations", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateReservationHandler)

	// 2. ACT
	handler.ServeHTTP(rr, req)

	// 3. ASSERT
	assert.Equal(t, http.StatusCreated, rr.Code, "El handler debería devolver código 201")
	expectedBody := `{"message":"Reserva creada exitosamente"}`
	assert.JSONEq(t, expectedBody, rr.Body.String(), "El JSON de respuesta no coincide")
}

// Test 2: Caso de Error (Método Incorrecto)
// Fíjate que esta función empieza FUERA de la llave de cierre de la anterior
func TestCreateReservationHandler_InvalidMethod(t *testing.T) {
	// 1. ARRANGE: Enviamos un GET en lugar de un POST
	req, err := http.NewRequest("GET", "/reservations", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateReservationHandler)

	// 2. ACT
	handler.ServeHTTP(rr, req)

	// 3. ASSERT: Esperamos un error 405 (Method Not Allowed)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Debería rechazar métodos que no sean POST")
}
