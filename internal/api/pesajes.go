package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type PesajesHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func (h *PesajesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *PesajesHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT p.id, p.guia, p.peso_bruto, p.tara, p.peso_neto, p.estado_balanza, p.fecha_creacion
		FROM pesajes p WHERE p.estado IN ('ACTIVO','INACTIVO') ORDER BY p.id DESC LIMIT 200`)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var guia sql.NullString
		var bruto, tara, neto sql.NullInt64
		var estadoBalanza, fecha sql.NullString
		_ = rows.Scan(&id, &guia, &bruto, &tara, &neto, &estadoBalanza, &fecha)
		data = append(data, map[string]any{
			"id": id, "guia": guia.String, "peso_bruto": bruto.Int64, "tara": tara.Int64,
			"peso_neto": neto.Int64, "estado_balanza": estadoBalanza.String, "fecha_creacion": fecha.String,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func (h *PesajesHandler) create(w http.ResponseWriter, r *http.Request) {
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

	clientEventID := strField(body, "client_event_id")
	if clientEventID != "" {
		var existID int64
		var guia sql.NullString
		err := h.DB.QueryRow(`SELECT id, guia FROM pesajes WHERE client_event_id = ?`, clientEventID).Scan(&existID, &guia)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true, "message": "Pesaje ya registrado", "id": existID,
				"guia": guia.String, "duplicado": true,
			})
			return
		}
	}

	bruto := intField(body, "peso_bruto")
	tara := intField(body, "tara")
	pesoNeto := 0
	if bruto > 0 && tara > 0 {
		pesoNeto = bruto - tara
		if pesoNeto < 0 {
			fail(w, http.StatusBadRequest, "El peso bruto no puede ser menor que la tara")
			return
		}
	}

	estadoBalanza := strField(body, "estado_balanza")
	if estadoBalanza != "FINALIZADO" {
		estadoBalanza = "EN_PROCESO"
	}
	estado := strField(body, "estado")
	if estado != "INACTIVO" {
		estado = "ACTIVO"
	}

	vehiculoID := nullIntField(body, "vehiculo_id")
	if estadoBalanza == "EN_PROCESO" && vehiculoID.Valid {
		var exist int64
		err := h.DB.QueryRow(`SELECT id FROM pesajes WHERE vehiculo_id = ? AND estado_balanza = 'EN_PROCESO' AND estado = 'ACTIVO' LIMIT 1`, vehiculoID.Int64).Scan(&exist)
		if err == nil {
			fail(w, http.StatusConflict, "Ya existe un pesaje en proceso para este vehículo")
			return
		}
	}

	guiaIn := strField(body, "guia")
	var guiaFinal sql.NullString
	if guiaIn != "" {
		var tomada int64
		err := h.DB.QueryRow(`SELECT id FROM pesajes WHERE guia = ?`, guiaIn).Scan(&tomada)
		if errors.Is(err, sql.ErrNoRows) {
			guiaFinal = sql.NullString{String: guiaIn, Valid: true}
		}
	}

	fecha := strField(body, "fecha_creacion")
	if fecha == "" {
		fecha = db.Ahora()
	}
	origen := strField(body, "origen")
	if origen == "" {
		origen = "WEB"
	}
	terminalID := strField(body, "terminal_id")

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
		INSERT INTO pesajes (
			guia, vehiculo_id, chofer_id, proveedor_id, producto_id, id_origen, id_destino,
			peso_bruto, tara, peso_neto, costo_bs, flete_bs_pendiente, pesador_id,
			estado, estado_balanza, client_event_id, origen, terminal_id,
			usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, sync_seq
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		nullStr(guiaFinal), vehiculoID, nullIntField(body, "chofer_id"), nullIntField(body, "proveedor_id"),
		nullIntField(body, "producto_id"), nullIntField(body, "id_origen"), nullIntField(body, "id_destino"),
		bruto, tara, pesoNeto, intField(body, "costo_bs"), intField(body, "flete_bs_pendiente"),
		nullIntField(body, "pesador_id"), estado, estadoBalanza, nullStrOr(clientEventID), origen, nullStrOr(terminalID),
		userID, userID, fecha, fecha, seq,
	)
	if err != nil {
		if clientEventID != "" && strings.Contains(err.Error(), "client_event_id") {
			var existID int64
			var guia sql.NullString
			if e := h.DB.QueryRow(`SELECT id, guia FROM pesajes WHERE client_event_id = ?`, clientEventID).Scan(&existID, &guia); e == nil {
				writeJSON(w, http.StatusOK, map[string]any{
					"success": true, "message": "Pesaje ya registrado", "id": existID, "guia": guia.String, "duplicado": true,
				})
				return
			}
		}
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	id, _ := res.LastInsertId()

	guiaAsignada := guiaFinal.String
	if !guiaFinal.Valid {
		guiaAsignada = fmt.Sprintf("PESO-%d-%06d", time.Now().Year(), id)
		_, _ = tx.Exec(`UPDATE pesajes SET guia = ? WHERE id = ?`, guiaAsignada, id)
	}
	if err := tx.Commit(); err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Pesaje registrado correctamente",
		"id":      id,
		"guia":    guiaAsignada,
		"data": map[string]any{"id": id, "guia": guiaAsignada},
	})
}

func strField(m map[string]any, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(v)
	}
}

func intField(m map[string]any, k string) int {
	v, ok := m[k]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	default:
		n, _ := strconv.Atoi(fmt.Sprint(v))
		return n
	}
}

func nullIntField(m map[string]any, k string) sql.NullInt64 {
	v, ok := m[k]
	if !ok || v == nil {
		return sql.NullInt64{}
	}
	n := intField(m, k)
	return sql.NullInt64{Int64: int64(n), Valid: n != 0 || v != nil}
}

func nullStr(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func nullStrOr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
