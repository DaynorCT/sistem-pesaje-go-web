package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type ChoferesHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *ChoferesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *ChoferesHandler) list(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	rows, err := h.DB.Query(`
		SELECT c.id, p.nombre, p.apellido_paterno, p.apellido_materno, p.ci, p.telefono, c.licencia, c.estado
		FROM choferes c JOIN personas p ON p.id = c.persona_id
		ORDER BY c.id DESC`)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var nombre, apPat, apMat, ci, tel, licencia, estado sql.NullString
		if err := rows.Scan(&id, &nombre, &apPat, &apMat, &ci, &tel, &licencia, &estado); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		data = append(data, map[string]any{
			"id": id, "nombre": nombre.String, "apellido_paterno": apPat.String,
			"apellido_materno": apMat.String, "ci": ci.String, "telefono": tel.String,
			"licencia": licencia.String, "estado": estado.String,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func (h *ChoferesHandler) create(w http.ResponseWriter, r *http.Request) {
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
	if strField(body, "nombre") == "" {
		fail(w, http.StatusBadRequest, "El nombre es requerido")
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
	res, err := tx.Exec(`INSERT INTO personas (nombre, apellido_paterno, apellido_materno, ci, telefono, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		strField(body, "nombre"), strField(body, "apellido_paterno"), strField(body, "apellido_materno"),
		strField(body, "ci"), strField(body, "telefono"), "ACTIVO", userID, userID, fecha, fecha)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	personaID, _ := res.LastInsertId()
	res, err = tx.Exec(`INSERT INTO choferes (persona_id, licencia, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,1,?)`,
		personaID, strField(body, "licencia"), "ACTIVO", userID, userID, fecha, fecha, seq)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Chofer creado correctamente", "id": id})
}

func (h *ChoferesHandler) update(w http.ResponseWriter, r *http.Request) {
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
	_, err = tx.Exec(`UPDATE personas SET nombre=?, apellido_paterno=?, apellido_materno=?, ci=?, telefono=?, usuario_modificacion=?, fecha_modificacion=? WHERE id = (SELECT persona_id FROM choferes WHERE id = ?)`,
		strField(body, "nombre"), strField(body, "apellido_paterno"), strField(body, "apellido_materno"),
		strField(body, "ci"), strField(body, "telefono"), userID, fecha, id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	estado := strField(body, "estado")
	if estado == "" {
		estado = "ACTIVO"
	}
	_, err = tx.Exec(`UPDATE choferes SET licencia=?, estado=?, usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`,
		strField(body, "licencia"), estado, userID, fecha, seq, id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Chofer actualizado correctamente"})
}
