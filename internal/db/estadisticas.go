package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ResumenInicio cifras del dashboard (paridad pesaje-go datos/estadisticas.go).
type ResumenInicio struct {
	PesajesTotal            int64      `json:"pesajesTotal"`
	PesajesHoy              int64      `json:"pesajesHoy"`
	PesajesEnProceso        int64      `json:"pesajesEnProceso"`
	PesajesFinalizadosHoy   int64      `json:"pesajesFinalizadosHoy"`
	VehiculosActivos        int64      `json:"vehiculosActivos"`
	ProductosActivos        int64      `json:"productosActivos"`
	ChoferesActivos         int64      `json:"choferesActivos"`
	ProveedoresActivos      int64      `json:"proveedoresActivos"`
	Ultimos7Dias            []PuntoDia `json:"ultimos7Dias"`
}

// PuntoDia un día para gráficos de inicio.
type PuntoDia struct {
	Fecha    string `json:"fecha"`
	Cantidad int64  `json:"cantidad"`
}

func CargarResumenInicio(ctx context.Context, conn *sql.DB) (ResumenInicio, error) {
	var r ResumenInicio
	var err error
	if err = conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM pesajes`).Scan(&r.PesajesTotal); err != nil {
		return r, err
	}
	if err = conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pesajes WHERE date(fecha_creacion) = date('now', 'localtime')`).Scan(&r.PesajesHoy); err != nil {
		return r, err
	}
	if err = conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pesajes WHERE estado = 'ACTIVO' AND estado_balanza = 'EN_PROCESO'`).Scan(&r.PesajesEnProceso); err != nil {
		return r, err
	}
	if err = conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pesajes WHERE estado_balanza = 'FINALIZADO' AND date(fecha_modificacion) = date('now', 'localtime')`).Scan(&r.PesajesFinalizadosHoy); err != nil {
		return r, err
	}
	if err = contarActivo(ctx, conn, `SELECT COUNT(*) FROM vehiculos WHERE estado = 'ACTIVO'`, &r.VehiculosActivos); err != nil {
		return r, err
	}
	if err = contarActivo(ctx, conn, `SELECT COUNT(*) FROM productos WHERE estado = 'ACTIVO'`, &r.ProductosActivos); err != nil {
		return r, err
	}
	if err = contarActivo(ctx, conn, `SELECT COUNT(*) FROM choferes WHERE estado = 'ACTIVO'`, &r.ChoferesActivos); err != nil {
		return r, err
	}
	if err = contarActivo(ctx, conn, `SELECT COUNT(*) FROM proveedores WHERE estado = 'ACTIVO'`, &r.ProveedoresActivos); err != nil {
		return r, err
	}
	r.Ultimos7Dias, err = cargarPesajesUltimosDias(ctx, conn, 7)
	return r, err
}

func contarActivo(ctx context.Context, conn *sql.DB, q string, dest *int64) error {
	return conn.QueryRowContext(ctx, q).Scan(dest)
}

func cargarPesajesUltimosDias(ctx context.Context, conn *sql.DB, dias int) ([]PuntoDia, error) {
	if dias <= 0 {
		dias = 7
	}
	q := fmt.Sprintf(`
SELECT date(fecha_creacion) AS d, COUNT(*) AS n
FROM pesajes
WHERE date(fecha_creacion) >= date('now', '-%d days', 'localtime')
GROUP BY d
ORDER BY d`, dias-1)
	rows, err := conn.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	porDia := map[string]int64{}
	for rows.Next() {
		var d string
		var n int64
		if err := rows.Scan(&d, &n); err != nil {
			return nil, err
		}
		porDia[d] = n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	loc := time.Local
	hoy := time.Now().In(loc)
	out := make([]PuntoDia, 0, dias)
	for i := dias - 1; i >= 0; i-- {
		t := hoy.AddDate(0, 0, -i)
		iso := t.Format("2006-01-02")
		out = append(out, PuntoDia{
			Fecha:    t.Format("02/01"),
			Cantidad: porDia[iso],
		})
	}
	return out, nil
}
