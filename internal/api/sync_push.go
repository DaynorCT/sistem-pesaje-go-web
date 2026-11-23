package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"sistema-pesaje-web/internal/auth"
	"sistema-pesaje-web/internal/db"
)

// postPush aplica cambios de catálogo enviados por el escritorio (no pesajes).
func (h *SyncHandler) postPush(w http.ResponseWriter, r *http.Request, recurso string) {
	if recurso == "pesajes" {
		fail(w, http.StatusBadRequest, "Use POST /api/pesajes para pesajes")
		return
	}
	_, ok := auth.UserIDFromRequest(r, h.JWTSecret)
	if !ok {
		fail(w, http.StatusUnauthorized, "No autorizado")
		return
	}
	meta, ok := catalogDefs[recurso]
	if !ok {
		fail(w, http.StatusNotFound, "Unknown resource: "+recurso)
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fail(w, http.StatusBadRequest, "Cuerpo inválido")
		return
	}
	clientEventID := strField(body, "client_event_id")
	if clientEventID == "" {
		fail(w, http.StatusBadRequest, "client_event_id is required")
		return
	}

	if prevID, res, found := h.operacionPrevia(clientEventID, recurso); found {
		fila := h.leerFilaCatalogo(recurso, prevID)
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true, "data": fila, "duplicado": true, "recurso": recurso, "resultado": res,
		})
		return
	}

	userID, _ := auth.UserIDFromRequest(r, h.JWTSecret)
	fecha := db.Ahora()
	id := int64(0)
	if v, ok := body["id"]; ok && v != nil {
		id, _ = strconv.ParseInt(fmt.Sprint(v), 10, 64)
	}
	versionBase := intField(body, "version_base")

	if id > 0 {
		var version int
		err := h.DB.QueryRow(fmt.Sprintf(`SELECT version FROM %s WHERE id = ?`, meta.tabla), id).Scan(&version)
		if err == sql.ErrNoRows {
			fail(w, http.StatusNotFound, "Registro no encontrado")
			return
		}
		if err != nil {
			fail(w, http.StatusInternalServerError, "Error del servidor")
			return
		}
		if versionBase < version {
			writeJSON(w, http.StatusConflict, map[string]any{
				"success":    false,
				"message":    "El servidor tiene una version mas nueva",
				"servidor":   h.leerFilaCatalogo(recurso, id),
			})
			return
		}
		if err := h.actualizarCatalogo(meta, recurso, userID, fecha, id, body); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = h.registrarOperacion(clientEventID, recurso, id, "updated")
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true, "data": h.leerFilaCatalogo(recurso, id), "recurso": recurso, "aplicado": "updated",
		})
		return
	}

	for _, req := range meta.requeridos {
		if strField(body, req) == "" {
			fail(w, http.StatusBadRequest, "Missing required field: "+req)
			return
		}
	}

	nuevoID, err := h.crearCatalogo(meta, recurso, userID, fecha, body)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = h.registrarOperacion(clientEventID, recurso, nuevoID, "created")
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "data": h.leerFilaCatalogo(recurso, nuevoID), "recurso": recurso, "aplicado": "created",
	})
}

type catalogDef struct {
	tabla      string
	persona    bool
	requeridos []string
}

var catalogDefs = map[string]catalogDef{
	"vehiculos":   {tabla: "vehiculos", requeridos: []string{"placa"}},
	"productos":   {tabla: "productos", requeridos: []string{"nombre"}},
	"origenes":    {tabla: "origenes", requeridos: []string{"nombre"}},
	"destinos":    {tabla: "destinos", requeridos: []string{"nombre"}},
	"choferes":    {tabla: "choferes", persona: true, requeridos: []string{"nombre"}},
	"proveedores": {tabla: "proveedores", persona: true, requeridos: []string{"nombre"}},
}

func (h *SyncHandler) operacionPrevia(clientEventID, recurso string) (entityID int64, resultado string, ok bool) {
	var ent sql.NullInt64
	err := h.DB.QueryRow(`SELECT entidad_id, resultado FROM sync_ops WHERE client_event_id = ?`, clientEventID).Scan(&ent, &resultado)
	if err != nil {
		return 0, "", false
	}
	return ent.Int64, resultado, ent.Valid
}

func (h *SyncHandler) registrarOperacion(clientEventID, recurso string, entidadID int64, resultado string) error {
	_, err := h.DB.Exec(
		`INSERT OR REPLACE INTO sync_ops (client_event_id, recurso, entidad_id, resultado, aplicado_en) VALUES (?,?,?,?,?)`,
		clientEventID, recurso, entidadID, resultado, db.Ahora(),
	)
	return err
}

func (h *SyncHandler) leerFilaCatalogo(recurso string, id int64) map[string]any {
	def, ok := recursos[recurso]
	if !ok || id <= 0 {
		return nil
	}
	// Delta query with since=-1 trick: use direct query by id from same SELECT shape
	switch recurso {
	case "vehiculos":
		return h.scanOne(`SELECT id, placa, tipo_vehiculo, tara, COALESCE(estado,'ACTIVO') estado, version, sync_seq, fecha_modificacion FROM vehiculos WHERE id = ?`, id)
	case "productos":
		return h.scanOne(`SELECT id, nombre, descripcion, COALESCE(estado,'ACTIVO') estado, version, sync_seq, fecha_modificacion FROM productos WHERE id = ?`, id)
	case "origenes":
		return h.scanOne(`SELECT id, nombre, descripcion, COALESCE(estado,'ACTIVO') estado, version, sync_seq, fecha_modificacion FROM origenes WHERE id = ?`, id)
	case "destinos":
		return h.scanOne(`SELECT id, nombre, descripcion, COALESCE(estado,'ACTIVO') estado, version, sync_seq, fecha_modificacion FROM destinos WHERE id = ?`, id)
	case "choferes":
		return h.scanOne(`SELECT c.id, p.nombre, p.apellido_paterno, p.apellido_materno, p.ci, p.telefono, c.licencia, c.estado, c.version, c.sync_seq, c.fecha_modificacion FROM choferes c JOIN personas p ON p.id = c.persona_id WHERE c.id = ?`, id)
	case "proveedores":
		return h.scanOne(`SELECT pr.id, pr.nombre_empresa, pr.descripcion, pr.estado, COALESCE(p.nombre,'') nombre, COALESCE(p.apellido_paterno,'') apellido_paterno, COALESCE(p.apellido_materno,'') apellido_materno, COALESCE(p.ci,'') ci, COALESCE(p.telefono,'') telefono, pr.version, pr.sync_seq, pr.fecha_modificacion FROM proveedores pr LEFT JOIN personas p ON p.id = pr.persona_id WHERE pr.id = ?`, id)
	default:
		_ = def
		return nil
	}
}

func (h *SyncHandler) scanOne(query string, id int64) map[string]any {
	rows, err := h.DB.Query(query, id)
	if err != nil {
		return nil
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	if !rows.Next() {
		return nil
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if rows.Scan(ptrs...) != nil {
		return nil
	}
	row := map[string]any{}
	for i, c := range cols {
		row[c] = jsonValue(vals[i])
	}
	return row
}

func (h *SyncHandler) crearCatalogo(meta catalogDef, recurso string, userID int64, fecha string, body map[string]any) (int64, error) {
	tx, err := h.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	seq, err := db.NextSeq(tx)
	if err != nil {
		return 0, err
	}
	var personaID sql.NullInt64
	if meta.persona {
		res, err := tx.Exec(`INSERT INTO personas (nombre, apellido_paterno, apellido_materno, ci, telefono, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			strField(body, "nombre"), strField(body, "apellido_paterno"), strField(body, "apellido_materno"),
			strField(body, "ci"), strField(body, "telefono"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha)
		if err != nil {
			return 0, err
		}
		pid, _ := res.LastInsertId()
		personaID = sql.NullInt64{Int64: pid, Valid: true}
	}

	var res sql.Result
	switch recurso {
	case "vehiculos":
		res, err = tx.Exec(`INSERT INTO vehiculos (placa, tipo_vehiculo, tara, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,?,1,?)`,
			strField(body, "placa"), strField(body, "tipo_vehiculo"), intField(body, "tara"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha, seq)
	case "productos":
		res, err = tx.Exec(`INSERT INTO productos (nombre, descripcion, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,1,?)`,
			strField(body, "nombre"), strField(body, "descripcion"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha, seq)
	case "origenes":
		res, err = tx.Exec(`INSERT INTO origenes (nombre, descripcion, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,1,?)`,
			strField(body, "nombre"), strField(body, "descripcion"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha, seq)
	case "destinos":
		res, err = tx.Exec(`INSERT INTO destinos (nombre, descripcion, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,1,?)`,
			strField(body, "nombre"), strField(body, "descripcion"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha, seq)
	case "choferes":
		res, err = tx.Exec(`INSERT INTO choferes (persona_id, licencia, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,1,?)`,
			personaID.Int64, strField(body, "licencia"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha, seq)
	case "proveedores":
		res, err = tx.Exec(`INSERT INTO proveedores (persona_id, nombre_empresa, descripcion, estado, usuario_creacion, usuario_modificacion, fecha_creacion, fecha_modificacion, version, sync_seq) VALUES (?,?,?,?,?,?,?,?,1,?)`,
			personaID.Int64, strField(body, "nombre_empresa"), strField(body, "descripcion"), strOr(body, "estado", "ACTIVO"), userID, userID, fecha, fecha, seq)
	default:
		return 0, fmt.Errorf("recurso no soportado")
	}
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (h *SyncHandler) actualizarCatalogo(meta catalogDef, recurso string, userID int64, fecha string, id int64, body map[string]any) error {
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	seq, err := db.NextSeq(tx)
	if err != nil {
		return err
	}
	if meta.persona {
		_, err = tx.Exec(`UPDATE personas SET
			nombre = COALESCE(?, nombre),
			apellido_paterno = COALESCE(?, apellido_paterno),
			apellido_materno = COALESCE(?, apellido_materno),
			ci = COALESCE(?, ci),
			telefono = COALESCE(?, telefono)
		WHERE id = (SELECT persona_id FROM `+meta.tabla+` WHERE id = ?)`,
			nulStrPtr(body, "nombre"), nulStrPtr(body, "apellido_paterno"), nulStrPtr(body, "apellido_materno"),
			nulStrPtr(body, "ci"), nulStrPtr(body, "telefono"), id)
		if err != nil {
			return err
		}
	}

	switch recurso {
	case "vehiculos":
		_, err = tx.Exec(`UPDATE vehiculos SET placa=COALESCE(?,placa), tipo_vehiculo=COALESCE(?,tipo_vehiculo), tara=COALESCE(?,tara), estado=COALESCE(?,estado), usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`,
			nulStrPtr(body, "placa"), nulStrPtr(body, "tipo_vehiculo"), nulIntPtr(body, "tara"), nulStrPtr(body, "estado"), userID, fecha, seq, id)
	case "productos":
		_, err = tx.Exec(`UPDATE productos SET nombre=COALESCE(?,nombre), descripcion=COALESCE(?,descripcion), estado=COALESCE(?,estado), usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`,
			nulStrPtr(body, "nombre"), nulStrPtr(body, "descripcion"), nulStrPtr(body, "estado"), userID, fecha, seq, id)
	case "origenes", "destinos":
		_, err = tx.Exec(fmt.Sprintf(`UPDATE %s SET nombre=COALESCE(?,nombre), descripcion=COALESCE(?,descripcion), estado=COALESCE(?,estado), usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`, meta.tabla),
			nulStrPtr(body, "nombre"), nulStrPtr(body, "descripcion"), nulStrPtr(body, "estado"), userID, fecha, seq, id)
	case "choferes":
		_, err = tx.Exec(`UPDATE choferes SET licencia=COALESCE(?,licencia), estado=COALESCE(?,estado), usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`,
			nulStrPtr(body, "licencia"), nulStrPtr(body, "estado"), userID, fecha, seq, id)
	case "proveedores":
		_, err = tx.Exec(`UPDATE proveedores SET nombre_empresa=COALESCE(?,nombre_empresa), descripcion=COALESCE(?,descripcion), estado=COALESCE(?,estado), usuario_modificacion=?, fecha_modificacion=?, version=version+1, sync_seq=? WHERE id=?`,
			nulStrPtr(body, "nombre_empresa"), nulStrPtr(body, "descripcion"), nulStrPtr(body, "estado"), userID, fecha, seq, id)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func strOr(body map[string]any, k, def string) string {
	s := strField(body, k)
	if s == "" {
		return def
	}
	return s
}

func nulStrPtr(body map[string]any, k string) any {
	if _, ok := body[k]; !ok {
		return nil
	}
	s := strField(body, k)
	if s == "" {
		return nil
	}
	return s
}

func nulIntPtr(body map[string]any, k string) any {
	if _, ok := body[k]; !ok {
		return nil
	}
	return intField(body, k)
}

