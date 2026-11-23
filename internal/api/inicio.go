package api

import (
	"database/sql"
	"net/http"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type InicioHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *InicioHandler) Resumen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	resumen, err := db.CargarResumenInicio(r.Context(), h.DB)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": resumen})
}
