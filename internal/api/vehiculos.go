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
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPut:
		h.update(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *VehiculosHandler) list(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	rows, err := h.DB.Query(`SELECT id, placa, COALESCE(tipo_vehiculo,''), COALESCE(tara,0), COALESCE(estado,'ACTIVO') FROM vehiculos ORDER BY id DESC`)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var placa, tipo, estado sql.NullString
		var tara int
		if err := rows.Scan(&id, &placa, &tipo, &tara, &estado); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		data = append(data, map[string]any{
			"id": id, "placa": placa.String, "tipo_vehiculo": tipo.String, "tara": tara, "estado": estado.String,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
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

func (h *VehiculosHandler) update(w http.ResponseWriter, r *http.Request) {
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
	id := intField(body, "id")
	if id <= 0 {
		fail(w, http.StatusBadRequest, "ID requerido")
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
	estado := strings.TrimSpace(strField(body, "estado"))
	if estado == "" {
		estado = "ACTIVO"
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
	_, err = tx.Exec(`UPDATE vehiculos SET placa=?, tipo_vehiculo=?, tara=?, estado=?, usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`,
		placa, strings.TrimSpace(strField(body, "tipo_vehiculo")), tara, estado, userID, fecha, seq, id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Vehículo actualizado correctamente"})
}
