package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/boleta"
)

type BoletaConfigHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *BoletaConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.put(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *BoletaConfigHandler) get(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	cfg, err := boleta.ObtenerConfig(r.Context(), h.DB)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": cfg})
}

func (h *BoletaConfigHandler) put(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromRequest(r, h.JWTSecret)
	if !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	var body boleta.Config
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	normalizarConfig(&body)
	if err := boleta.GuardarConfig(r.Context(), h.DB, body, userID); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	cfg, _ := boleta.ObtenerConfig(r.Context(), h.DB)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Configuración de boleta guardada.",
		"data":    cfg,
	})
}

func normalizarConfig(c *boleta.Config) {
	c.Salida = strings.TrimSpace(strings.ToUpper(c.Salida))
	c.Direccion = strings.TrimSpace(strings.ToUpper(c.Direccion))
	c.Celular1 = strings.TrimSpace(c.Celular1)
	c.Celular2 = strings.TrimSpace(c.Celular2)
	c.Ciudad = strings.TrimSpace(strings.ToUpper(c.Ciudad))
	c.TituloSuperior = strings.TrimSpace(strings.ToUpper(c.TituloSuperior))
	c.Marca = strings.TrimSpace(strings.ToUpper(c.Marca))
	c.TituloDocumento = strings.TrimSpace(strings.ToUpper(c.TituloDocumento))
	c.Acreditacion = strings.TrimSpace(strings.ToUpper(c.Acreditacion))
}

func (h *PesajesHandler) BoletaPDF(w http.ResponseWriter, r *http.Request) {
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
	if strings.ToUpper(strings.TrimSpace(estadoBalanza)) != "FINALIZADO" {
		fail(w, http.StatusConflict, "La boleta solo está disponible para pesajes FINALIZADOS")
		return
	}
	d, err := boleta.CargarDatos(r.Context(), h.DB, id)
	if err != nil {
		fail(w, http.StatusNotFound, err.Error())
		return
	}
	cfg, err := boleta.ObtenerConfig(r.Context(), h.DB)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	pdf, err := boleta.GenerarPDF(d, cfg)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	guia := d.Guia
	if guia == "" {
		guia = fmt.Sprintf("%d", id)
	}
	filename := fmt.Sprintf("boleta-%s.pdf", guia)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}
