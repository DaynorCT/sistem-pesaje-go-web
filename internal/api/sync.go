package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

type SyncHandler struct {
	DB        *sql.DB
	JWTSecret string
}

type recursoSync struct {
	nombre string
	query  string
}

var recursos = map[string]recursoSync{
	"vehiculos": {
		"vehiculos",
		`SELECT id, placa, tipo_vehiculo, tara, COALESCE(estado,'ACTIVO') as estado, version, sync_seq, fecha_modificacion FROM vehiculos WHERE sync_seq > ? ORDER BY sync_seq ASC`,
	},
	"productos": {
		"productos",
		`SELECT id, nombre, descripcion, COALESCE(estado,'ACTIVO') as estado, version, sync_seq, fecha_modificacion FROM productos WHERE sync_seq > ? ORDER BY sync_seq ASC`,
	},
	"origenes": {
		"origenes",
		`SELECT id, nombre, descripcion, COALESCE(estado,'ACTIVO') as estado, version, sync_seq, fecha_modificacion FROM origenes WHERE sync_seq > ? ORDER BY sync_seq ASC`,
	},
	"destinos": {
		"destinos",
		`SELECT id, nombre, descripcion, COALESCE(estado,'ACTIVO') as estado, version, sync_seq, fecha_modificacion FROM destinos WHERE sync_seq > ? ORDER BY sync_seq ASC`,
	},
	"choferes": {
		"choferes",
		`SELECT c.id, p.nombre, p.apellido_paterno, p.apellido_materno, p.ci, p.telefono, c.licencia, c.estado, c.version, c.sync_seq, c.fecha_modificacion
		FROM choferes c JOIN personas p ON p.id = c.persona_id WHERE c.sync_seq > ? ORDER BY c.sync_seq ASC`,
	},
	"proveedores": {
		"proveedores",
		`SELECT pr.id, pr.nombre_empresa, pr.descripcion, pr.estado, COALESCE(p.nombre,'') nombre, COALESCE(p.apellido_paterno,'') apellido_paterno,
		COALESCE(p.apellido_materno,'') apellido_materno, COALESCE(p.ci,'') ci, COALESCE(p.telefono,'') telefono, pr.version, pr.sync_seq, pr.fecha_modificacion
		FROM proveedores pr LEFT JOIN personas p ON p.id = pr.persona_id WHERE pr.sync_seq > ? ORDER BY pr.sync_seq ASC`,
	},
	"pesajes": {
		"pesajes",
		`SELECT p.id, p.guia, p.vehiculo_id, p.chofer_id, p.proveedor_id, p.producto_id, p.id_origen, p.id_destino,
		p.peso_bruto, p.tara, p.peso_neto, p.costo_bs, p.flete_bs_pendiente, p.estado, p.estado_balanza, p.fecha_creacion,
		p.version, p.sync_seq, p.fecha_modificacion
		FROM pesajes p WHERE p.sync_seq > ? ORDER BY p.sync_seq ASC`,
	},
}

func (h *SyncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sync/")
	path = strings.Trim(path, "/")
	if path == "" {
		fail(w, http.StatusNotFound, "Unknown resource")
		return
	}
	recurso := strings.ToLower(path)
	if r.Method == http.MethodPost {
		h.postPush(w, r, recurso)
		return
	}
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	if _, ok := auth.UserIDFromRequest(r, h.JWTSecret); !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	def, ok := recursos[recurso]
	if !ok {
		fail(w, http.StatusNotFound, "Unknown resource: "+path)
		return
	}
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
	if since < 0 {
		since = 0
	}
	rows, err := h.DB.Query(def.query, since)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var data []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		row := map[string]any{}
		for i, c := range cols {
			row[c] = jsonValue(vals[i])
		}
		data = append(data, row)
	}
	cursor, _ := db.SecuenciaActual(h.DB)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"cursor":  cursor,
		"recurso": def.nombre,
	})
}

func jsonValue(v any) any {
	switch t := v.(type) {
	case []byte:
		return string(t)
	default:
		return t
	}
}

