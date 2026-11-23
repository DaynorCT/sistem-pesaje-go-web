package db

import (
	"context"
	"database/sql"
	"strings"
)

// FilaPesaje fila para grilla/reportes (paridad pesaje-go internal/datos.FilaPesaje).
type FilaPesaje struct {
	ID               int64
	Guia             string
	NombreChofer     string
	PlacaVehiculo    string
	LicenciaChofer   string
	TipoVehiculo     string
	NombreProveedor  string
	NombreProducto   string
	Origen           string
	Destino          string
	CostoBs          sql.NullInt64
	FleteBsPendiente sql.NullInt64
	FechaCreacion    string
	PesoBruto        sql.NullInt64
	Tara             sql.NullInt64
	PesoNeto         sql.NullInt64
	Estado           string
	EstadoBalanza    string
}

func (f FilaPesaje) Fecha() string { return soloFecha(f.FechaCreacion) }
func (f FilaPesaje) Hora() string  { return soloHora(f.FechaCreacion) }

func soloFecha(fechaCompleta string) string {
	if len(fechaCompleta) >= 10 {
		return fechaCompleta[8:10] + "/" + fechaCompleta[5:7] + "/" + fechaCompleta[0:4]
	}
	return ""
}

func soloHora(fechaCompleta string) string {
	if len(fechaCompleta) >= 16 {
		return fechaCompleta[11:16]
	}
	return ""
}

const listarPesajesSelect = `
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
LEFT JOIN destinos d ON d.id = p.id_destino`

// ListarPesajesFiltrado listado para reportes (solo ACTIVO; mismos criterios que el PDF).
func ListarPesajesFiltrado(ctx context.Context, conn *sql.DB, desde, hasta, placa string, limite int) ([]FilaPesaje, error) {
	if limite <= 0 {
		limite = 500
	}
	q := strings.Builder{}
	q.WriteString(listarPesajesSelect)
	q.WriteString(` WHERE p.estado = 'ACTIVO'`)
	args := []any{}
	if strings.TrimSpace(desde) != "" {
		q.WriteString(` AND date(p.fecha_creacion) >= date(?)`)
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		q.WriteString(` AND date(p.fecha_creacion) <= date(?)`)
		args = append(args, strings.TrimSpace(hasta))
	}
	if strings.TrimSpace(placa) != "" {
		q.WriteString(` AND UPPER(v.placa) LIKE ?`)
		args = append(args, "%"+strings.ToUpper(strings.TrimSpace(placa))+"%")
	}
	q.WriteString(` ORDER BY p.id DESC LIMIT ?`)
	args = append(args, limite)
	rows, err := conn.QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return escanearFilasPesaje(rows)
}

func escanearFilasPesaje(rows *sql.Rows) ([]FilaPesaje, error) {
	var out []FilaPesaje
	for rows.Next() {
		var f FilaPesaje
		var guia sql.NullString
		if err := rows.Scan(
			&f.ID, &guia, &f.NombreChofer, &f.PlacaVehiculo, &f.LicenciaChofer, &f.TipoVehiculo,
			&f.NombreProveedor, &f.NombreProducto, &f.Origen, &f.Destino,
			&f.CostoBs, &f.FleteBsPendiente, &f.FechaCreacion,
			&f.PesoBruto, &f.Tara, &f.PesoNeto, &f.Estado, &f.EstadoBalanza,
		); err != nil {
			return nil, err
		}
		f.Guia = guia.String
		out = append(out, f)
	}
	return out, rows.Err()
}

// NombreEmpresaActiva para encabezado del reporte PDF.
func NombreEmpresaActiva(ctx context.Context, conn *sql.DB) string {
	var n sql.NullString
	_ = conn.QueryRowContext(ctx,
		`SELECT nombre_empresa FROM empresas WHERE estado = 'ACTIVO' ORDER BY id DESC LIMIT 1`).Scan(&n)
	return n.String
}
