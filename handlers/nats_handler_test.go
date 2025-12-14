package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"ms3/models"

	"github.com/glebarez/sqlite"
	natssrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Función auxiliar para crear una DB de prueba rápida
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Falló conexión DB: %v", err)
	}
	err = db.AutoMigrate(&models.Reservation{})
	if err != nil {
		t.Fatalf("Falló automigrate: %v", err)
	}
	return db
}

func Test_unwrapNestEnvelope_WhenNestEnvelope_ReturnsInnerData(t *testing.T) {
	payload := []byte(`{"pattern":"reservation.create","data":{"user_id":"u1"},"id":"req-1"}`)

	reqID, inner := unwrapNestEnvelope(payload)

	assert.Equal(t, "req-1", reqID)
	assert.JSONEq(t, `{"user_id":"u1"}`, string(inner))
}

func Test_unwrapNestEnvelope_WhenNotNest_ReturnsOriginal(t *testing.T) {
	payload := []byte(`{"user_id":"u1"}`)

	reqID, inner := unwrapNestEnvelope(payload)

	assert.Equal(t, "", reqID)
	assert.JSONEq(t, `{"user_id":"u1"}`, string(inner))
}

func Test_decodeString_JSONString(t *testing.T) {
	assert.Equal(t, "abc", decodeString([]byte(`"abc"`)))
}

func Test_decodeString_Raw(t *testing.T) {
	assert.Equal(t, "xyz", decodeString([]byte(`xyz`)))
}

func TestProcessCreateReservation_WrappedFormat(t *testing.T) {
	db := setupTestDB(t)

	// Simulamos el JSON envuelto
	inputJSON := `{
		"data": {
			"user_id": "user-123",
			"car_id": "car-99",
			"date": "2025-10-10"
		},
		"id": "msg-id"
	}`

	responseBytes := ProcessCreateReservation(db, []byte(inputJSON))

	var respRes models.Reservation
	err := json.Unmarshal(responseBytes, &respRes)
	assert.NoError(t, err)
	assert.Equal(t, "PENDIENTE", respRes.Status)
	assert.Equal(t, "user-123", respRes.UserID)
}

func TestProcessCreateReservation_DirectFormat(t *testing.T) {
	db := setupTestDB(t)

	// Formato directo (sin wrapper "data")
	inputJSON := `{"user_id":"user-777","car_id":"car-1","date":"2025-10-11"}`

	responseBytes := ProcessCreateReservation(db, []byte(inputJSON))

	var respRes models.Reservation
	err := json.Unmarshal(responseBytes, &respRes)
	assert.NoError(t, err)
	assert.Equal(t, "PENDIENTE", respRes.Status)
	assert.Equal(t, "user-777", respRes.UserID)
}

func TestProcessCreateReservation_BadJSON_ReturnsFormatoIncorrecto(t *testing.T) {
	db := setupTestDB(t)

	responseBytes := ProcessCreateReservation(db, []byte(`{bad json`))
	assert.Equal(t, `{"error":"Formato incorrecto"}`, string(responseBytes))
}

func TestProcessGetReservations(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.Reservation{UserID: "test-user", Status: "CONFIRMADA"})
	responseBytes := ProcessGetReservations(db)

	var list []models.Reservation
	_ = json.Unmarshal(responseBytes, &list)

	assert.NotEmpty(t, list)
	assert.Equal(t, "test-user", list[0].UserID)
}

func TestProcessGetPendingReservations_FiltersOnlyPendiente(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.Reservation{UserID: "u1", Status: "PENDIENTE"})
	db.Create(&models.Reservation{UserID: "u1", Status: "ACEPTADA"})

	responseBytes := ProcessGetPendingReservations(db)

	var list []models.Reservation
	_ = json.Unmarshal(responseBytes, &list)

	assert.Len(t, list, 1)
	assert.Equal(t, "PENDIENTE", list[0].Status)
}

func TestProcessGetClientReservations_FiltersByUser(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.Reservation{UserID: "u1", Status: "PENDIENTE"})
	db.Create(&models.Reservation{UserID: "u2", Status: "PENDIENTE"})

	responseBytes := ProcessGetClientReservations(db, "u2")

	var list []models.Reservation
	_ = json.Unmarshal(responseBytes, &list)

	assert.Len(t, list, 1)
	assert.Equal(t, "u2", list[0].UserID)
}

func TestProcessUpdateReservation_BadJSON_ReturnsError(t *testing.T) {
	db := setupTestDB(t)

	out := ProcessUpdateReservation(db, []byte(`{"data":{`)) // JSON roto
	assert.Equal(t, `{"error":"Formato JSON incorrecto"}`, string(out))
}

func TestProcessUpdateReservation_NotFound_ReturnsError(t *testing.T) {
	db := setupTestDB(t)

	payload := `{"data":{"id":999,"status":"ACEPTADA","mechanic_id":"m1"}}`
	out := ProcessUpdateReservation(db, []byte(payload))

	assert.Equal(t, `{"error":"Reserva no encontrada"}`, string(out))
}

func TestProcessUpdateReservation_Success_UpdatesStatusAndMechanic(t *testing.T) {
	db := setupTestDB(t)

	// crea una reserva inicial
	r := models.Reservation{UserID: "u1", Status: "PENDIENTE"}
	err := db.Create(&r).Error
	assert.NoError(t, err)

	payload := map[string]any{
		"data": map[string]any{
			"id":          r.ID,
			"status":      "ACEPTADA",
			"mechanic_id": "m1",
		},
	}
	b, _ := json.Marshal(payload)

	out := ProcessUpdateReservation(db, b)

	var updated models.Reservation
	err = json.Unmarshal(out, &updated)
	assert.NoError(t, err)

	assert.Equal(t, "ACEPTADA", updated.Status)
	// mechanic_id en tu código es *string, así que checkea pointer
	if assert.NotNil(t, updated.MechanicID) {
		assert.Equal(t, "m1", *updated.MechanicID)
	}
}

func TestProcessCreateReservation_DBError(t *testing.T) {
	db := setupTestDB(t)

	// Provocamos error en Create: eliminamos la tabla
	_ = db.Migrator().DropTable(&models.Reservation{})

	inputJSON := `{"user_id":"user-err","car_id":"c1","date":"2025-10-10"}`
	out := ProcessCreateReservation(db, []byte(inputJSON))

	assert.Equal(t, `{"error":"No se pudo guardar en BD"}`, string(out))
}

func setupNATS(t *testing.T) (*natssrv.Server, *nats.Conn) {
	t.Helper()

	opts := &natssrv.Options{
		Port:   -1, // puerto random
		NoLog:  true,
		NoSigs: true,
	}
	s, err := natssrv.NewServer(opts)
	if err != nil {
		t.Fatalf("no pude crear nats server: %v", err)
	}
	go s.Start()

	if !s.ReadyForConnections(5 * time.Second) {
		s.Shutdown()
		t.Fatalf("nats no quedó ready")
	}

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		s.Shutdown()
		t.Fatalf("no pude conectar a nats: %v", err)
	}

	return s, nc
}

type nestReplyMsg struct {
	ID         string          `json:"id"`
	Response   json.RawMessage `json:"response"`
	IsDisposed bool            `json:"isDisposed"`
}

func Test_CreateReservation_NATS_DirectPayload_RespondsRaw(t *testing.T) {
	db := setupTestDB(t) // <-- usa la tuya existente
	s, nc := setupNATS(t)
	defer nc.Close()
	defer s.Shutdown()

	CreateReservation(nc, db)
	_ = nc.Flush()

	payload := []byte(`{"user_id":"u1","car_id":"c1","date":"2025-10-10"}`)

	msg, err := nc.Request("reservation.create", payload, 2*time.Second)
	assert.NoError(t, err)

	var r models.Reservation
	err = json.Unmarshal(msg.Data, &r)
	assert.NoError(t, err)
	assert.Equal(t, "u1", r.UserID)
	assert.Equal(t, "PENDIENTE", r.Status)
}

func Test_CreateReservation_NATS_Envelope_RespondsNestFormat(t *testing.T) {
	db := setupTestDB(t)
	s, nc := setupNATS(t)
	defer nc.Close()
	defer s.Shutdown()

	CreateReservation(nc, db)
	_ = nc.Flush()

	payload := []byte(`{
		"pattern":"reservation.create",
		"data":{"user_id":"u2","car_id":"c2","date":"2025-10-11"},
		"id":"req-123"
	}`)

	msg, err := nc.Request("reservation.create", payload, 2*time.Second)
	assert.NoError(t, err)

	var out nestReplyMsg
	err = json.Unmarshal(msg.Data, &out)
	assert.NoError(t, err)
	assert.Equal(t, "req-123", out.ID)
	assert.True(t, out.IsDisposed)

	var r models.Reservation
	err = json.Unmarshal(out.Response, &r)
	assert.NoError(t, err)
	assert.Equal(t, "u2", r.UserID)
	assert.Equal(t, "PENDIENTE", r.Status)
}

func Test_GetClientReservations_NATS_UsesDecodeString(t *testing.T) {
	db := setupTestDB(t)
	s, nc := setupNATS(t)
	defer nc.Close()
	defer s.Shutdown()

	_ = db.Create(&models.Reservation{UserID: "uA", Status: "PENDIENTE"}).Error
	_ = db.Create(&models.Reservation{UserID: "uB", Status: "PENDIENTE"}).Error

	GetClientReservations(nc, db)
	_ = nc.Flush()

	msg1, err := nc.Request("reservation.get_by_user", []byte(`"uA"`), 2*time.Second)
	assert.NoError(t, err)
	var list1 []models.Reservation
	_ = json.Unmarshal(msg1.Data, &list1)
	assert.Len(t, list1, 1)
	assert.Equal(t, "uA", list1[0].UserID)

	msg2, err := nc.Request("reservation.get_by_user", []byte(`uB`), 2*time.Second)
	assert.NoError(t, err)
	var list2 []models.Reservation
	_ = json.Unmarshal(msg2.Data, &list2)
	assert.Len(t, list2, 1)
	assert.Equal(t, "uB", list2[0].UserID)
}
