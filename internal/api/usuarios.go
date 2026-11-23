package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/clave"
	"sistema-pesaje-web/internal/db"
)

type UsuariosHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *UsuariosHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPut:
		h.update(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *UsuariosHandler) list(w http.ResponseWriter) {
	rows, err := h.DB.Query(`
		SELECT u.id, p.nombre, p.apellido_paterno, p.apellido_materno, p.ci, u.email, u.rol, p.telefono, u.estado
		FROM usuarios u JOIN personas p ON p.id = u.persona_id
		ORDER BY u.id DESC`)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var nom, apP, apM, ci, email, rol, tel, estado sql.NullString
		if err := rows.Scan(&id, &nom, &apP, &apM, &ci, &email, &rol, &tel, &estado); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		nc := strings.TrimSpace(strings.Join([]string{nom.String, apP.String, apM.String}, " "))
		data = append(data, map[string]any{
			"id": id, "nombre": nom.String, "apellido_paterno": apP.String, "apellido_materno": apM.String,
			"ci": ci.String, "correo": email.String, "rol": rol.String, "telefono": tel.String,
			"estado": estado.String, "nombre_completo": nc,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func (h *UsuariosHandler) create(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	nombre := strings.TrimSpace(strField(body, "nombre"))
	correo := strings.TrimSpace(strField(body, "correo"))
	if nombre == "" {
		fail(w, http.StatusBadRequest, "El nombre es requerido")
		return
	}
	if correo == "" {
		fail(w, http.StatusBadRequest, "El correo es requerido")
		return
	}
	clavePlana := strField(body, "clave")
	if clavePlana == "" {
		clavePlana = strField(body, "contrasena")
	}
	if clavePlana == "" {
		fail(w, http.StatusBadRequest, "La contraseña es requerida")
		return
	}
	rol := strings.TrimSpace(strField(body, "rol"))
	if rol == "" {
		rol = "usuario"
	}
	fecha := db.Ahora()
	hash := clave.Hash(clavePlana)

	tx, err := h.DB.Begin()
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		INSERT INTO personas (nombre, apellido_paterno, apellido_materno, ci, correo, telefono, estado, fecha_creacion, fecha_modificacion)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		nombre, strField(body, "apellido_paterno"), strField(body, "apellido_materno"),
		strField(body, "ci"), correo, strField(body, "telefono"), "ACTIVO", fecha, fecha,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			fail(w, http.StatusConflict, "Correo o documento ya registrado")
			return
		}
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	personaID, _ := res.LastInsertId()
	res, err = tx.Exec(`
		INSERT INTO usuarios (persona_id, email, password_hash, rol, estado, fecha_creacion, fecha_modificacion)
		VALUES (?,?,?,?,?,?,?)`,
		personaID, correo, hash, rol, "ACTIVO", fecha, fecha,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			fail(w, http.StatusConflict, "El correo ya está en uso")
			return
		}
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "message": "Usuario creado correctamente", "id": id,
	})
}

func (h *UsuariosHandler) update(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
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
	correo := strings.TrimSpace(strField(body, "correo"))
	if nombre == "" || correo == "" {
		fail(w, http.StatusBadRequest, "Nombre y correo son requeridos")
		return
	}
	rol := strings.TrimSpace(strField(body, "rol"))
	if rol == "" {
		rol = "usuario"
	}
	estado := strings.TrimSpace(strField(body, "estado"))
	if estado == "" {
		estado = "ACTIVO"
	}
	clavePlana := strField(body, "clave")
	if clavePlana == "" {
		clavePlana = strField(body, "contrasena")
	}
	fecha := db.Ahora()

	tx, err := h.DB.Begin()
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer func() { _ = tx.Rollback() }()

	var personaID int64
	if err := tx.QueryRow(`SELECT persona_id FROM usuarios WHERE id = ?`, id).Scan(&personaID); err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "Usuario no encontrado")
		return
	} else if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}

	_, err = tx.Exec(`
		UPDATE personas SET nombre=?, apellido_paterno=?, apellido_materno=?, ci=?, correo=?, telefono=?, estado=?, fecha_modificacion=?
		WHERE id = ?`,
		nombre, strField(body, "apellido_paterno"), strField(body, "apellido_materno"),
		strField(body, "ci"), correo, strField(body, "telefono"), estado, fecha, personaID,
	)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}

	if clavePlana != "" {
		hash := clave.Hash(clavePlana)
		_, err = tx.Exec(`
			UPDATE usuarios SET email=?, rol=?, estado=?, password_hash=?, fecha_modificacion=? WHERE id=?`,
			correo, rol, estado, hash, fecha, id)
	} else {
		_, err = tx.Exec(`
			UPDATE usuarios SET email=?, rol=?, estado=?, fecha_modificacion=? WHERE id=?`,
			correo, rol, estado, fecha, id)
	}
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			fail(w, http.StatusConflict, "El correo ya está en uso")
			return
		}
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": fmt.Sprintf("Usuario #%d actualizado", id)})
}