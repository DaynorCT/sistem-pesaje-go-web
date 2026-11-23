package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

func pesajeIDFromRequest(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *PesajesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	id, ok := pesajeIDFromRequest(r)
	if !ok {
		fail(w, http.StatusBadRequest, "ID inválido")
		return
	}
	row, err := h.scanPesaje(id)
	if err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "Pesaje no encontrado")
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": row})
}

func (h *PesajesHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	userID, ok := auth.UserIDFromRequest(r, h.JWTSecret)
	if !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	id, ok := pesajeIDFromRequest(r)
	if !ok {
		fail(w, http.StatusBadRequest, "ID inválido")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}

	var estadoBalanza string
	err := h.DB.QueryRow(`SELECT estado_balanza FROM pesajes WHERE id = ?`, id).Scan(&estadoBalanza)
	if err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "Pesaje no encontrado")
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	if estadoBalanza != "EN_PROCESO" {
		fail(w, http.StatusConflict, "Solo se puede editar un pesaje EN PROCESO")
		return
	}

	bruto := intField(body, "peso_bruto")
	tara := intField(body, "tara")
	if bruto > 0 && tara > 0 && bruto <= tara {
		fail(w, http.StatusBadRequest, "El peso bruto debe ser mayor que la tara")
		return
	}
	pesoNeto := 0
	if bruto > 0 && tara > 0 {
		pesoNeto = bruto - tara
	}
	guia := strField(body, "guia")
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
	res, err := tx.Exec(`
		UPDATE pesajes SET
			guia = CASE WHEN ? != '' THEN ? ELSE guia END,
			peso_bruto = CASE WHEN ? > 0 THEN ? ELSE peso_bruto END,
			tara = CASE WHEN ? > 0 THEN ? ELSE tara END,
			peso_neto = CASE WHEN ? > 0 AND ? > 0 THEN ? ELSE peso_neto END,
			usuario_modificacion = ?, fecha_modificacion = ?, sync_seq = ?
		WHERE id = ? AND estado_balanza = 'EN_PROCESO'`,
		guia, guia, bruto, bruto, tara, tara, bruto, tara, pesoNeto,
		userID, fecha, seq, id,
	)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fail(w, http.StatusConflict, "No se pudo actualizar el pesaje")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	row, _ := h.scanPesaje(id)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Pesaje actualizado", "data": row})
}

func (h *PesajesHandler) FinalizarByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	userID, ok := auth.UserIDFromRequest(r, h.JWTSecret)
	if !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	id, ok := pesajeIDFromRequest(r)
	if !ok {
		fail(w, http.StatusBadRequest, "ID inválido")
		return
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
	res, err := tx.Exec(`
		UPDATE pesajes SET
			estado_balanza = 'FINALIZADO',
			peso_neto = COALESCE(peso_bruto, 0) - COALESCE(tara, 0),
			usuario_modificacion = ?, fecha_modificacion = ?, sync_seq = ?
		WHERE id = ? AND estado_balanza = 'EN_PROCESO'
		  AND COALESCE(peso_bruto, 0) > 0
		  AND COALESCE(tara, 0) > 0
		  AND COALESCE(peso_bruto, 0) > COALESCE(tara, 0)`,
		userID, fecha, seq, id,
	)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fail(w, http.StatusConflict,
			"No se puede finalizar: debe estar EN PROCESO con bruto y tara mayores a cero y bruto mayor que la tara")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	row, _ := h.scanPesaje(id)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Pesaje finalizado", "data": row})
}

func (h *PesajesHandler) scanPesaje(id int64) (map[string]any, error) {
	var guia sql.NullString
	var bruto, tara, neto sql.NullInt64
	var estadoBalanza, estado, fecha sql.NullString
	err := h.DB.QueryRow(`
		SELECT guia, peso_bruto, tara, peso_neto, estado_balanza, estado, fecha_creacion
		FROM pesajes WHERE id = ?`, id).Scan(&guia, &bruto, &tara, &neto, &estadoBalanza, &estado, &fecha)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":             id,
		"guia":           guia.String,
		"peso_bruto":     bruto.Int64,
		"tara":           tara.Int64,
		"peso_neto":      neto.Int64,
		"estado_balanza": estadoBalanza.String,
		"estado":         estado.String,
		"fecha_creacion": fecha.String,
	}, nil
}