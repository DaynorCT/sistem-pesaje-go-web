package boleta_test

import (
	"testing"

	"sistema-pesaje-web/internal/boleta"
)

func TestGenerarPDFBoleta(t *testing.T) {
	d := &boleta.Datos{
		ID: 1, Guia: "G-1", Fecha: "22/11/2026", Hora: "10:00",
		VehiculoPlaca: "ABC123", PesoBruto: 1000, Tara: 200, PesoNeto: 800,
	}
	cfg := &boleta.Config{Marca: "TEST", TituloSuperior: "BOLETA"}
	b, err := boleta.GenerarPDF(d, cfg, nil)
	if err != nil || len(b) < 4 || string(b[:4]) != "%PDF" {
		t.Fatalf("pdf invalido: len=%d err=%v", len(b), err)
	}
	if len(boleta.LogoSistemaJPEG()) == 0 {
		t.Fatal("logo no embebido")
	}
}
