package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type UltimoPesoHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *UltimoPesoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w)
	case http.MethodPost:
		h.post(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *UltimoPesoHandler) get(w http.ResponseWriter) {
	var peso float64
	var fecha sql.NullString
	err := h.DB.QueryRow(`SELECT peso, fecha FROM peso_en_vivo WHERE id = 1`).Scan(&peso, &fecha)
	if err == sql.ErrNoRows {
		peso = 0
	} else if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"peso":    peso,
		"fecha":   nullFecha(fecha),
	})
}

func (h *UltimoPesoHandler) post(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	var body struct {
		Peso float64 `json:"peso"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	if body.Peso < 0 || body.Peso != body.Peso {
		fail(w, http.StatusBadRequest, "Peso invalido")
		return
	}
	fecha := db.Ahora()
	_, err := h.DB.Exec(`
		INSERT INTO peso_en_vivo (id, peso, fecha) VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET peso = excluded.peso, fecha = excluded.fecha`,
		body.Peso, fecha)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Peso en vivo actualizado"})
}

func nullFecha(f sql.NullString) any {
	if f.Valid {
		return f.String
	}
	return nil
}
