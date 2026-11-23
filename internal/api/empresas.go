package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type EmpresasHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *EmpresasHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPut:
		h.update(w, r)
	case http.MethodDelete:
		h.toggleEstado(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *EmpresasHandler) list(w http.ResponseWriter) {
	rows, err := h.DB.Query(`
		SELECT id, nombre_empresa, actividad_economica, correo_electronico, telefono, logo, estado
		FROM empresas ORDER BY id DESC`)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var nombre, actividad, correo, telefono, logo, estado sql.NullString
		if err := rows.Scan(&id, &nombre, &actividad, &correo, &telefono, &logo, &estado); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		data = append(data, map[string]any{
			"id": id, "nombre_empresa": nombre.String, "actividad_economica": actividad.String,
			"correo_electronico": correo.String, "telefono": telefono.String, "logo": logo.String, "estado": estado.String,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func (h *EmpresasHandler) create(w http.ResponseWriter, r *http.Request) {
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
	nombre := strings.TrimSpace(strField(body, "nombre_empresa"))
	actividad := strings.TrimSpace(strField(body, "actividad_economica"))
	if nombre == "" {
		fail(w, http.StatusBadRequest, "El nombre de la empresa es obligatorio")
		return
	}
	fecha := db.Ahora()
	_, err := h.DB.Exec(`
		INSERT INTO empresas (nombre_empresa, actividad_economica, correo_electronico, telefono, logo, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion)
		VALUES (?, ?, ?, ?, ?, 'ACTIVO', ?, ?, ?, ?)`,
		nombre, actividad, strings.TrimSpace(strField(body, "correo_electronico")), strings.TrimSpace(strField(body, "telefono")), strField(body, "logo"),
		userID, userID, fecha, fecha)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Empresa creada correctamente"})
}

func (h *EmpresasHandler) update(w http.ResponseWriter, r *http.Request) {
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
	nombre := strings.TrimSpace(strField(body, "nombre_empresa"))
	if id <= 0 || nombre == "" {
		fail(w, http.StatusBadRequest, "ID y nombre son obligatorios")
		return
	}
	estado := strings.TrimSpace(strField(body, "estado"))
	if estado == "" {
		estado = "ACTIVO"
	}
	if estado != "ACTIVO" && estado != "INACTIVO" {
		fail(w, http.StatusBadRequest, "Estado inválido")
		return
	}
	logo := strField(body, "logo")
	fecha := db.Ahora()
	_, err := h.DB.Exec(`
		UPDATE empresas SET
			nombre_empresa = ?,
			actividad_economica = ?,
			correo_electronico = ?,
			telefono = ?,
			estado = ?,
			logo = CASE WHEN ? != '' THEN ? ELSE logo END,
			usuario_modificacion = ?,
			fecha_modificacion = ?
		WHERE id = ?`,
		nombre,
		strings.TrimSpace(strField(body, "actividad_economica")),
		strings.TrimSpace(strField(body, "correo_electronico")),
		strings.TrimSpace(strField(body, "telefono")),
		estado,
		logo, logo,
		userID, fecha, id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Empresa actualizada correctamente"})
}

func (h *EmpresasHandler) toggleEstado(w http.ResponseWriter, r *http.Request) {
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
	estado := "INACTIVO"
	if body["activar"] == true {
		estado = "ACTIVO"
	}
	_, err := h.DB.Exec(`UPDATE empresas SET estado = ?, usuario_modificacion = ?, fecha_modificacion = ? WHERE id = ?`,
		estado, userID, db.Ahora(), id)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Empresa actualizada correctamente"})
}

func (h *EmpresasHandler) Branding(w http.ResponseWriter, _ *http.Request) {
	var nombre, logo sql.NullString
	err := h.DB.QueryRow(`SELECT nombre_empresa, logo FROM empresas WHERE estado = 'ACTIVO' ORDER BY id DESC LIMIT 1`).Scan(&nombre, &logo)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    map[string]any{"nombre": "Sistema de Pesaje", "logo": ""},
		})
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	n := "Sistema de Pesaje"
	if nombre.Valid && nombre.String != "" {
		n = nombre.String
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"nombre": n, "logo": logo.String},
	})
}
