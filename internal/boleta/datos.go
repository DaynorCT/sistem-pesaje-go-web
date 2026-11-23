package boleta

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Datos para PDF (paridad pesaje-go internal/datos/datos_boleta.go).
type Datos struct {
	ID              int64
	Guia            string
	Fecha           string
	Hora            string
	ChoferNombre    string
	ChoferCI        string
	ChoferLicencia  string
	VehiculoPlaca   string
	VehiculoTipo    string
	ProveedorNombre string
	ProductoNombre  string
	OrigenNombre    string
	DestinoNombre   string
	Operador        string
	PesoBruto       int
	Tara            int
	PesoNeto        int
	CostoBs         int
	FleteBs         int
}

const selectDatosBoleta = `
SELECT p.id, p.guia, p.fecha_creacion,
       TRIM(COALESCE(pe.nombre, '') || ' ' || COALESCE(pe.apellido_paterno, '')
            || ' ' || COALESCE(pe.apellido_materno, '')) AS chofer_nombre,
       COALESCE(pe.ci, '') AS chofer_ci,
       COALESCE(ch.licencia, '') AS chofer_licencia,
       COALESCE(v.placa, '') AS vehiculo_placa,
       COALESCE(v.tipo_vehiculo, '') AS vehiculo_tipo,
       COALESCE(NULLIF(TRIM(COALESCE(pper.nombre, '') || ' '
            || COALESCE(pper.apellido_paterno, '') || ' '
            || COALESCE(pper.apellido_materno, '')), ''), pr.nombre_empresa, '') AS proveedor_nombre,
       COALESCE(prod.nombre, '') AS producto_nombre,
       COALESCE(o.nombre, '') AS origen_nombre,
       COALESCE(d.nombre, '') AS destino_nombre,
       COALESCE(ps.nombre, '') AS operador,
       COALESCE(p.peso_bruto, 0), COALESCE(p.tara, 0), COALESCE(p.peso_neto, 0),
       COALESCE(p.costo_bs, 0), COALESCE(p.flete_bs_pendiente, 0)
FROM pesajes p
LEFT JOIN vehiculos v ON v.id = p.vehiculo_id
LEFT JOIN choferes ch ON ch.id = p.chofer_id
LEFT JOIN personas pe ON pe.id = ch.persona_id
LEFT JOIN proveedores pr ON pr.id = p.proveedor_id
LEFT JOIN personas pper ON pper.id = pr.persona_id
LEFT JOIN productos prod ON prod.id = p.producto_id
LEFT JOIN origenes o ON o.id = p.id_origen
LEFT JOIN destinos d ON d.id = p.id_destino
LEFT JOIN personas ps ON ps.id = p.pesador_id`

func CargarDatos(ctx context.Context, conn *sql.DB, pesajeID int64) (*Datos, error) {
	q := selectDatosBoleta + ` WHERE p.id = ?`
	var d Datos
	var guia sql.NullString
	var fc string
	err := conn.QueryRowContext(ctx, q, pesajeID).Scan(
		&d.ID, &guia, &fc,
		&d.ChoferNombre, &d.ChoferCI, &d.ChoferLicencia,
		&d.VehiculoPlaca, &d.VehiculoTipo, &d.ProveedorNombre,
		&d.ProductoNombre, &d.OrigenNombre, &d.DestinoNombre, &d.Operador,
		&d.PesoBruto, &d.Tara, &d.PesoNeto, &d.CostoBs, &d.FleteBs,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pesaje no encontrado")
	}
	if err != nil {
		return nil, err
	}
	d.Guia = guia.String
	d.Fecha, d.Hora = partesFechaHora(fc)
	return &d, nil
}

// CargarDatosReporte filas para PDF de reporte (solo pesajes ACTIVO; filtros opcionales).
func CargarDatosReporte(ctx context.Context, conn *sql.DB, desde, hasta, placa string) ([]Datos, error) {
	q := strings.Builder{}
	q.WriteString(selectDatosBoleta)
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
	q.WriteString(` ORDER BY p.id DESC`)
	rows, err := conn.QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Datos
	for rows.Next() {
		d, err := scanDatosBoletaRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func scanDatosBoletaRow(rows *sql.Rows) (*Datos, error) {
	var d Datos
	var guia sql.NullString
	var fc string
	err := rows.Scan(
		&d.ID, &guia, &fc,
		&d.ChoferNombre, &d.ChoferCI, &d.ChoferLicencia,
		&d.VehiculoPlaca, &d.VehiculoTipo, &d.ProveedorNombre,
		&d.ProductoNombre, &d.OrigenNombre, &d.DestinoNombre, &d.Operador,
		&d.PesoBruto, &d.Tara, &d.PesoNeto, &d.CostoBs, &d.FleteBs,
	)
	if err != nil {
		return nil, err
	}
	d.Guia = guia.String
	d.Fecha, d.Hora = partesFechaHora(fc)
	return &d, nil
}

func partesFechaHora(fechaCompleta string) (fecha, hora string) {
	if len(fechaCompleta) >= 16 {
		fecha = fechaCompleta[8:10] + "/" + fechaCompleta[5:7] + "/" + fechaCompleta[0:4]
		hora = fechaCompleta[11:16]
	}
	return fecha, hora
}
