package handlers

import (
	"testing"
)

func TestNewReservationHandler(t *testing.T) {
	// Verificamos que el constructor funcione.
	// Pasamos 'nil' como base de datos porque en esta prueba simple
	// solo queremos ver si se crea el objeto struct, no vamos a conectar a la BD real.
	h := NewReservationHandler(nil)

	if h == nil {
		t.Error("El handler no se creó correctamente (es nil)")
	}
}
