package reportes_test

import (
	"testing"

	"sistema-pesaje-web/internal/reportes"
)

func TestGenerarReporteProducePDF(t *testing.T) {
	b, err := reportes.GenerarPDF(nil, "Empresa")
	if err != nil || len(b) < 4 || string(b[:4]) != "%PDF" {
		t.Fatalf("pdf reporte invalido: len=%d err=%v", len(b), err)
	}
}
