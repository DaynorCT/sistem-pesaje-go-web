package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/clave"
)

type AuthHandlers struct {
	DB           *sql.DB
	JWTSecret    string
	RefreshSecret string
	CookieSecure bool
}

func (h *AuthHandlers) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	var body struct {
		Usuario    string `json:"usuario"`
		Contrasena string `json:"contrasena"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	if body.Usuario == "" || body.Contrasena == "" {
		fail(w, http.StatusBadRequest, "Usuario y contraseña requeridos")
		return
	}
	u, err := loadUser(h.DB, body.Usuario)
	if err != nil || u == nil {
		fail(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}
	if !clave.Verificar(body.Contrasena, u.PasswordHash) {
		fail(w, http.StatusUnauthorized, "Contraseña incorrecta")
		return
	}
	access, err := auth.CreateAccessToken(h.JWTSecret, u.ID, u.Email, u.Rol)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	refresh, err := auth.CreateRefreshToken(h.RefreshSecret, u.ID)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success":       true,
		"message":       "Token generado correctamente",
		"access_token":  access,
		"refresh_token": refresh,
		"user": map[string]any{
			"id": u.ID, "email": u.Email, "rol": u.Rol,
		},
	})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	var body struct {
		Usuario    string `json:"usuario"`
		Contrasena string `json:"contrasena"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	if body.Usuario == "" || body.Contrasena == "" {
		fail(w, http.StatusBadRequest, "Usuario y contraseña requeridos")
		return
	}
	u, err := loadUser(h.DB, body.Usuario)
	if err != nil || u == nil {
		fail(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}
	if !clave.Verificar(body.Contrasena, u.PasswordHash) {
		fail(w, http.StatusUnauthorized, "Contraseña incorrecta")
		return
	}
	access, _ := auth.CreateAccessToken(h.JWTSecret, u.ID, u.Email, u.Rol)
	refresh, _ := auth.CreateRefreshToken(h.RefreshSecret, u.ID)
	auth.SetAuthCookies(w, access, refresh, h.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Login exitoso",
		"user": map[string]any{
			"id": u.ID, "email": u.Email, "rol": u.Rol, "nombre": u.DisplayName(),
		},
	})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	refresh := auth.CookieValue(r, "refresh_token")
	if refresh == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Token de Refresh requerido"})
		return
	}
	userID, err := auth.VerifyRefresh(h.RefreshSecret, refresh)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Token de Refresh inválido"})
		return
	}
	u, err := loadUserByID(h.DB, userID)
	if err != nil || u == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Usuario no encontrado o inactivo"})
		return
	}
	access, err := auth.CreateAccessToken(h.JWTSecret, u.ID, u.Email, u.Rol)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "access_token", Value: access, Path: "/", HttpOnly: true,
		Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode, MaxAge: 15 * 60,
	})
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Token refrescado", "access_token": access})
}

type usuarioRow struct {
	ID           int64
	Email        string
	PasswordHash string
	Rol          string
	Nombre       sql.NullString
	ApPat        sql.NullString
	ApMat        sql.NullString
}

func (u *usuarioRow) DisplayName() string {
	parts := []string{}
	if u.Nombre.Valid && u.Nombre.String != "" {
		parts = append(parts, u.Nombre.String)
	}
	if u.ApPat.Valid && u.ApPat.String != "" {
		parts = append(parts, u.ApPat.String)
	}
	if u.ApMat.Valid && u.ApMat.String != "" {
		parts = append(parts, u.ApMat.String)
	}
	if len(parts) == 0 {
		return u.Email
	}
	return strings.Join(parts, " ")
}

func loadUser(db *sql.DB, email string) (*usuarioRow, error) {
	var u usuarioRow
	err := db.QueryRow(`
		SELECT u.id, u.email, u.password_hash, COALESCE(u.rol,'operador'), p.nombre, p.apellido_paterno, p.apellido_materno
		FROM usuarios u
		LEFT JOIN personas p ON p.id = u.persona_id
		WHERE u.email = ? AND u.estado = 'ACTIVO'`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Rol, &u.Nombre, &u.ApPat, &u.ApMat,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if u.Rol == "" {
		u.Rol = "operador"
	}
	return &u, nil
}

func loadUserByID(db *sql.DB, id int64) (*usuarioRow, error) {
	var u usuarioRow
	err := db.QueryRow(`
		SELECT u.id, u.email, u.password_hash, COALESCE(u.rol,'operador'), p.nombre, p.apellido_paterno, p.apellido_materno
		FROM usuarios u
		LEFT JOIN personas p ON p.id = u.persona_id
		WHERE u.id = ? AND u.estado = 'ACTIVO'`, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Rol, &u.Nombre, &u.ApPat, &u.ApMat,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}
