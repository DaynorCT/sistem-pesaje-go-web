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

	vehiculoID := intField(body, "vehiculo_id")
	if vehiculoID <= 0 {
		fail(w, http.StatusBadRequest, "Seleccione un vehículo (placa).")
		return
	}
	bruto := intField(body, "peso_bruto")
	tara := intField(body, "tara")
	if bruto <= 0 || tara <= 0 || bruto <= tara {
		fail(w, http.StatusBadRequest, "El peso bruto debe ser mayor que la tara.")
		return
	}
	pesoNeto := bruto - tara
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
			vehiculo_id = ?, chofer_id = ?, proveedor_id = ?, producto_id = ?,
			id_origen = ?, id_destino = ?,
			peso_bruto = ?, tara = ?, peso_neto = ?,
			costo_bs = ?, flete_bs_pendiente = ?,
			usuario_modificacion = ?, fecha_modificacion = ?, sync_seq = ?
		WHERE id = ? AND estado_balanza = 'EN_PROCESO'`,
		vehiculoID,
		nullIntField(body, "chofer_id"), nullIntField(body, "proveedor_id"),
		nullIntField(body, "producto_id"), nullIntField(body, "id_origen"), nullIntField(body, "id_destino"),
		bruto, tara, pesoNeto,
		intField(body, "costo_bs"), intField(body, "flete_bs_pendiente"),
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
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Pesaje actualizado correctamente", "data": row})
}

func (h *PesajesHandler) CambiarEstadoByID(w http.ResponseWriter, r *http.Request) {
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
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	activo, _ := body["activo"].(bool)
	estado := "INACTIVO"
	if activo {
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
	res, err := tx.Exec(`
		UPDATE pesajes SET estado = ?, usuario_modificacion = ?, fecha_modificacion = ?, sync_seq = ?
		WHERE id = ?`, estado, userID, fecha, seq, id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fail(w, http.StatusNotFound, "Pesaje no encontrado")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Estado actualizado"})
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
	var vehiculo, chofer, proveedor, producto, origen, destino sql.NullInt64
	var bruto, tara, neto, costo, flete sql.NullInt64
	var estadoBalanza, estado, fecha sql.NullString
	err := h.DB.QueryRow(`
		SELECT vehiculo_id, chofer_id, proveedor_id, producto_id, id_origen, id_destino,
			guia, peso_bruto, tara, peso_neto, costo_bs, flete_bs_pendiente,
			estado_balanza, estado, fecha_creacion
		FROM pesajes WHERE id = ?`, id).Scan(
		&vehiculo, &chofer, &proveedor, &producto, &origen, &destino,
		&guia, &bruto, &tara, &neto, &costo, &flete,
		&estadoBalanza, &estado, &fecha,
	)
	if err != nil {
		return nil, err
	}
	row := map[string]any{
		"id":                 id,
		"vehiculo_id":        vehiculo.Int64,
		"guia":               guia.String,
		"peso_bruto":         bruto.Int64,
		"tara":               tara.Int64,
		"peso_neto":          neto.Int64,
		"costo_bs":           costo.Int64,
		"flete_bs_pendiente": flete.Int64,
		"estado_balanza":     estadoBalanza.String,
		"estado":             estado.String,
		"fecha_creacion":     fecha.String,
	}
	if chofer.Valid {
		row["chofer_id"] = chofer.Int64
	}
	if proveedor.Valid {
		row["proveedor_id"] = proveedor.Int64
	}
	if producto.Valid {
		row["producto_id"] = producto.Int64
	}
	if origen.Valid {
		row["id_origen"] = origen.Int64
	}
	if destino.Valid {
		row["id_destino"] = destino.Int64
	}
	return row, nil
}
