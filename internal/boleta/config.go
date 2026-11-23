package boleta

import (
	"context"
	"database/sql"

	"sistema-pesaje-web/internal/db"
)

// Config fila boleta_config (paridad pesaje-go).
type Config struct {
	ID              int64  `json:"id"`
	Salida          string `json:"salida"`
	Direccion       string `json:"direccion"`
	Celular1        string `json:"celular1"`
	Celular2        string `json:"celular2"`
	Ciudad          string `json:"ciudad"`
	TituloSuperior  string `json:"tituloSuperior"`
	Marca           string `json:"marca"`
	TituloDocumento string `json:"tituloDocumento"`
	Acreditacion    string `json:"acreditacion"`
}

func ObtenerConfig(ctx context.Context, conn *sql.DB) (*Config, error) {
	row := conn.QueryRowContext(ctx, `
SELECT id, salida, direccion, celular1, celular2, ciudad, titulo_superior, marca, titulo_documento, acreditacion
FROM boleta_config ORDER BY id DESC LIMIT 1`)
	var b Config
	err := row.Scan(&b.ID, &b.Salida, &b.Direccion, &b.Celular1, &b.Celular2, &b.Ciudad,
		&b.TituloSuperior, &b.Marca, &b.TituloDocumento, &b.Acreditacion)
	if err == sql.ErrNoRows {
		return &Config{
			Salida: "Salida a Potosi", Direccion: "Tarija km 2",
			Celular1: "2782323", Celular2: "1234343",
			Ciudad: "POTOSI - BOLIVIA",
			TituloSuperior: "BALANZA DE PESAJE DIGITAL",
			Marca: "PRIMAVERA",
			TituloDocumento: "BOLETA DE PESAJE",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GuardarConfig(ctx context.Context, conn *sql.DB, b Config, usuarioID int64) error {
	fecha := db.Ahora()
	if b.ID <= 0 {
		act, err := ObtenerConfig(ctx, conn)
		if err != nil {
			return err
		}
		if act != nil && act.ID > 0 {
			b.ID = act.ID
		}
	}
	if b.ID <= 0 {
		_, err := conn.ExecContext(ctx, `
INSERT INTO boleta_config (salida, direccion, celular1, celular2, ciudad, titulo_superior, marca,
 titulo_documento, acreditacion, usuario_modificacion, fecha_modificacion)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			b.Salida, b.Direccion, b.Celular1, b.Celular2, b.Ciudad,
			b.TituloSuperior, b.Marca, b.TituloDocumento, b.Acreditacion, usuarioID, fecha)
		return err
	}
	_, err := conn.ExecContext(ctx, `
UPDATE boleta_config SET salida=?, direccion=?, celular1=?, celular2=?, ciudad=?, titulo_superior=?,
 marca=?, titulo_documento=?, acreditacion=?, usuario_modificacion=?, fecha_modificacion=? WHERE id=?`,
		b.Salida, b.Direccion, b.Celular1, b.Celular2, b.Ciudad,
		b.TituloSuperior, b.Marca, b.TituloDocumento, b.Acreditacion, usuarioID, fecha, b.ID)
	return err
}
