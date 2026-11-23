// Boleta de ejemplo: go run ./cmd/boleta-muestra
package main

import (
	"os"
	"path/filepath"

	"sistema-pesaje-web/internal/boleta"
)

func main() {
	d := &boleta.Datos{
		ID: 42, Guia: "G-2026-001", Fecha: "23/11/2026", Hora: "14:30",
		ChoferNombre: "Juan Perez", ChoferLicencia: "LP-12345",
		VehiculoPlaca: "ABC-1234", VehiculoTipo: "Camion",
		ProveedorNombre: "Agro Primavera S.R.L.",
		ProductoNombre: "Soja", OrigenNombre: "Campo Norte", DestinoNombre: "Silo Central",
		Operador: "Maria Lopez",
		PesoBruto: 45230, Tara: 15200, PesoNeto: 30030,
		CostoBs: 150, FleteBs: 80,
	}
	cfg := &boleta.Config{
		Salida: "Salida a Potosi", Direccion: "Tarija km 2",
		Celular1: "2782323", Celular2: "70123456",
		Ciudad: "Potosi - Bolivia",
		TituloSuperior: "BALANZA DE PESAJE DIGITAL",
		Marca: "PRIMAVERA",
		TituloDocumento: "BOLETA DE PESAJE",
		Acreditacion: "SENASAG",
	}
	b, err := boleta.GenerarPDF(d, cfg, nil)
	if err != nil {
		panic(err)
	}
	out := filepath.Join("design", "boleta-ejemplo-nuevo.pdf")
	if err := os.WriteFile(out, b, 0o644); err != nil {
		panic(err)
	}
}
