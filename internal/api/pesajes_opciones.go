package api

import (
	"database/sql"
	"net/http"
	"strconv"
)

const listarPesajesGrillaSQL = `
SELECT p.id, p.guia,
       TRIM(COALESCE(cper.nombre, '') || ' ' || COALESCE(cper.apellido_paterno, '')
            || ' ' || COALESCE(cper.apellido_materno, '')) AS chofer,
       COALESCE(v.placa, '') AS placa, COALESCE(c.licencia, '') AS licencia,
       COALESCE(v.tipo_vehiculo, '') AS tipo,
       COALESCE(NULLIF(TRIM(COALESCE(pper.nombre, '') || ' '
            || COALESCE(pper.apellido_paterno, '') || ' '
            || COALESCE(pper.apellido_materno, '')), ''), pr.nombre_empresa, '') AS proveedor,
       COALESCE(prod.nombre, '') AS producto,
       COALESCE(o.nombre, '') AS origen, COALESCE(d.nombre, '') AS destino,
       p.costo_bs, p.flete_bs_pendiente, p.fecha_creacion,
       p.peso_bruto, p.tara, p.peso_neto, p.estado, p.estado_balanza
FROM pesajes p
LEFT JOIN vehiculos v ON v.id = p.vehiculo_id
LEFT JOIN choferes c ON c.id = p.chofer_id
LEFT JOIN personas cper ON cper.id = c.persona_id
LEFT JOIN proveedores pr ON pr.id = p.proveedor_id
LEFT JOIN personas pper ON pper.id = pr.persona_id
LEFT JOIN productos prod ON prod.id = p.producto_id
LEFT JOIN origenes o ON o.id = p.id_origen
LEFT JOIN destinos d ON d.id = p.id_destino
WHERE p.estado IN ('ACTIVO', 'INACTIVO')
ORDER BY p.id DESC LIMIT 80`

func (h *PesajesHandler) Opciones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	opciones := map[string]any{
		"vehiculos":   queryOpciones(h.DB, `SELECT id, placa FROM vehiculos WHERE estado='ACTIVO' ORDER BY placa`),
		"choferes":    queryOpciones(h.DB, `SELECT c.id, TRIM(COALESCE(p.nombre,'')||' '||COALESCE(p.apellido_paterno,'')) FROM choferes c JOIN personas p ON p.id=c.persona_id WHERE c.estado='ACTIVO' ORDER BY p.nombre`),
		"proveedores": queryOpciones(h.DB, `SELECT pr.id, COALESCE(NULLIF(pr.nombre_empresa,''), TRIM(p.nombre||' '||COALESCE(p.apellido_paterno,''))) FROM proveedores pr LEFT JOIN personas p ON p.id=pr.persona_id WHERE pr.estado='ACTIVO' ORDER BY 2`),
		"productos":   queryOpciones(h.DB, `SELECT id, nombre FROM productos WHERE estado='ACTIVO' ORDER BY nombre`),
		"origenes":    queryOpciones(h.DB, `SELECT id, nombre FROM origenes WHERE estado='ACTIVO' ORDER BY nombre`),
		"destinos":    queryOpciones(h.DB, `SELECT id, nombre FROM destinos WHERE estado='ACTIVO' ORDER BY nombre`),
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": opciones})
}

func (h *PesajesHandler) Grilla(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	rows, err := h.DB.Query(listarPesajesGrillaSQL)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var guia, chofer, placa, lic, tipo, prov, prod, orig, dest, fecha, est, estBal sql.NullString
		var costo, flete, bruto, tara, neto sql.NullInt64
		if err := rows.Scan(&id, &guia, &chofer, &placa, &lic, &tipo, &prov, &prod, &orig, &dest,
			&costo, &flete, &fecha, &bruto, &tara, &neto, &est, &estBal); err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		data = append(data, map[string]any{
			"id": id, "guia": guia.String,
			"nombre_chofer": chofer.String, "placa": placa.String, "licencia": lic.String,
			"tipo_vehiculo": tipo.String, "nombre_proveedor": prov.String, "producto": prod.String,
			"origen": orig.String, "destino": dest.String,
			"costo_bs": costo.Int64, "flete_bs_pendiente": flete.Int64, "fecha_creacion": fecha.String,
			"peso_bruto": bruto.Int64, "tara": tara.Int64, "peso_neto": neto.Int64,
			"estado": est.String, "estado_balanza": estBal.String,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func (h *PesajesHandler) LookupVehiculo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if id <= 0 {
		fail(w, http.StatusBadRequest, "ID inválido")
		return
	}
	var tipo sql.NullString
	var tara sql.NullInt64
	err := h.DB.QueryRow(`SELECT tipo_vehiculo, COALESCE(tara,0) FROM vehiculos WHERE id = ?`, id).Scan(&tipo, &tara)
	if err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "Vehículo no encontrado")
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"tipo": tipo.String, "tara": tara.Int64},
	})
}

func (h *PesajesHandler) LookupChofer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if id <= 0 {
		fail(w, http.StatusBadRequest, "ID inválido")
		return
	}
	var lic sql.NullString
	err := h.DB.QueryRow(`SELECT licencia FROM choferes WHERE id = ?`, id).Scan(&lic)
	if err == sql.ErrNoRows {
		fail(w, http.StatusNotFound, "Chofer no encontrado")
		return
	}
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error del servidor")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"licencia": lic.String},
	})
}

func queryOpciones(db *sql.DB, q string) []map[string]any {
	rows, err := db.Query(q)
	if err != nil {
		return []map[string]any{}
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var label sql.NullString
		if err := rows.Scan(&id, &label); err != nil {
			continue
		}
		out = append(out, map[string]any{"id": id, "label": label.String})
	}
	if out == nil {
		out = []map[string]any{}
	}
	return out
}