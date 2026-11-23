package db_test

import (
	"context"
	"path/filepath"
	"testing"

	"sistema-pesaje-web/internal/db"
)

func TestCargarResumenInicioVacío(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	res, err := db.CargarResumenInicio(context.Background(), sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if res.PesajesTotal != 0 || len(res.Ultimos7Dias) != 7 {
		t.Fatalf("resumen inesperado: %+v", res)
	}
}
