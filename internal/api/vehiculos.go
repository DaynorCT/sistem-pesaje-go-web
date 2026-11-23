package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type VehiculosHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *VehiculosHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.create(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *VehiculosHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromRequest(r, h.JWTSecret)
	if !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	placa := strings.TrimSpace(strField(body, "placa"))
	if placa == "" {
		fail(w, http.StatusBadRequest, "La placa es obligatoria")
		return
	}
	tara := intField(body, "tara")
	if tara < 0 {
		tara = 0
	}
	fecha := db.Ahora()
	tx, err := h.DB.Begin()
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer func() { _ = tx.Rollback() }()
	seq, err := db.NextSeq(tx)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	res, err := tx.Exec(`INSERT INTO vehiculos (placa, tipo_vehiculo, tara, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,?,1,?)`,
		placa, strings.TrimSpace(strField(body, "tipo_vehiculo")), tara, "ACTIVO", userID, userID, fecha, fecha, seq)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Vehículo creado correctamente", "id": id})
}