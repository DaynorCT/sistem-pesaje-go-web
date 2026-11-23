package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

// CatalogoHandler ABM para tablas nombre + descripcion (productos, origenes, destinos).
type CatalogoHandler struct {
	DB        *sql.DB
	JWTSecret string
	Tabla     string
	Label     string
}

func (h *CatalogoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *CatalogoHandler) list(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	q := fmt.Sprintf(`SELECT id, nombre, COALESCE(descripcion,''), COALESCE(estado,'ACTIVO') FROM %s ORDER BY id DESC`, h.Tabla)
	rows, err := h.DB.Query(q)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var nombre, desc, estado sql.NullString
		if err := rows.Scan(&id, &nombre, &desc, &estado); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		data = append(data, map[string]any{
			"id": id, "nombre": nombre.String, "descripcion": desc.String, "estado": estado.String,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func (h *CatalogoHandler) create(w http.ResponseWriter, r *http.Request) {
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
	nombre := strings.TrimSpace(strField(body, "nombre"))
	if nombre == "" {
		fail(w, http.StatusBadRequest, "El nombre es requerido")
		return
	}
	desc := strings.TrimSpace(strField(body, "descripcion"))
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
	ins := fmt.Sprintf(`INSERT INTO %s (nombre, descripcion, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,1,?)`, h.Tabla)
	res, err := tx.Exec(ins, nombre, desc, "ACTIVO", userID, userID, fecha, fecha, seq)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	msg := h.Label + " creado correctamente"
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg, "id": id})
}

func (h *CatalogoHandler) update(w http.ResponseWriter, r *http.Request) {
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
	nombre := strings.TrimSpace(strField(body, "nombre"))
	if nombre == "" {
		fail(w, http.StatusBadRequest, "El nombre es requerido")
		return
	}
	desc := strings.TrimSpace(strField(body, "descripcion"))
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
	up := fmt.Sprintf(`UPDATE %s SET nombre=?, descripcion=?, estado=?, usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`, h.Tabla)
	_, err = tx.Exec(up, nombre, desc, estado, userID, fecha, seq, id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": h.Label + " actualizado correctamente"})
}
