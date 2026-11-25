package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/boleta"
	"sistema-pesaje-web/internal/db"
	"sistema-pesaje-web/internal/reportes"
)

type ReportesHandler struct {
	DB        *sql.DB
	JWTSecret string
}

type filaReporteJSON struct {
	ID              int64  `json:"id"`
	Guia            string `json:"guia"`
	NombreChofer    string `json:"nombreChofer"`
	Placa           string `json:"placa"`
	NombreProveedor string `json:"nombreProveedor"`
	PesoBruto       int    `json:"pesoBruto"`
	Tara            int    `json:"tara"`
	PesoNeto        int    `json:"pesoNeto"`
	EstadoBalanza   string `json:"estadoBalanza"`
	Fecha           string `json:"fecha"`
	Hora            string `json:"hora"`
}

func (h *ReportesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listar(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *ReportesHandler) filtroDesdeQuery(r *http.Request) (desde, hasta, placa string) {
	q := r.URL.Query()
	desde = strings.TrimSpace(q.Get("desde"))
	hasta = strings.TrimSpace(q.Get("hasta"))
	placa = strings.TrimSpace(q.Get("placa"))
	return desde, hasta, placa
}

func (h *ReportesHandler) listar(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	desde, hasta, placa := h.filtroDesdeQuery(r)
	limite := 500
	if raw := strings.TrimSpace(r.URL.Query().Get("limite")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limite = n
		}
	}
	filas, err := db.ListarPesajesFiltrado(r.Context(), h.DB, desde, hasta, placa, limite)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	out := make([]filaReporteJSON, len(filas))
	for i, f := range filas {
		out[i] = filaReporteJSON{
			ID:              f.ID,
			Guia:            f.Guia,
			NombreChofer:    f.NombreChofer,
			Placa:           f.PlacaVehiculo,
			NombreProveedor: f.NombreProveedor,
			PesoBruto:       nullInt64(f.PesoBruto),
			Tara:            nullInt64(f.Tara),
			PesoNeto:        nullInt64(f.PesoNeto),
			EstadoBalanza:   f.EstadoBalanza,
			Fecha:           f.Fecha(),
			Hora:            f.Hora(),
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": out})
}

func (h *ReportesHandler) PDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	desde, hasta, placa := h.filtroDesdeQuery(r)
	items, err := boleta.CargarDatosReporte(r.Context(), h.DB, desde, hasta, placa)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	empresa := db.NombreEmpresaActiva(r.Context(), h.DB)
	pdf, err := reportes.GenerarPDF(items, reportes.Encabezado{
		Empresa: empresa,
		Desde:   desde,
		Hasta:   hasta,
		Placa:   placa,
	})
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	nombre := fmt.Sprintf("reporte-pesaje-%s.pdf", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, nombre))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func nullInt64(v sql.NullInt64) int {
	if !v.Valid {
		return 0
	}
	return int(v.Int64)
}
