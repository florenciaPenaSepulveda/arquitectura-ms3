package handlers

import (
	"encoding/json"
	"net/http"
)

// CreateReservationHandler simula la respuesta de crear una reserva
func CreateReservationHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Aquí iría la lógica de llamar a la base de datos (lo haremos después)
	// AGREGA ESTO AL PRINCIPIO:
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// ... el resto de tu código sigue igual ...
	// 2. Por ahora, respondemos que todo salió bien (Status 201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// 3. Enviamos un mensaje de respuesta
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Reserva creada exitosamente",
	})
}
