// Boleta de ejemplo: go run ./cmd/boleta-muestra
package main

import (
	"os"
	"path/filepath"

	"sistema-pesaje-web/internal/boleta"
)

func main() {
	d := &boleta.Datos{
		ID: 1, Guia: "PESO-2026-000001", Fecha: "23/11/2026", Hora: "15:37",
		ChoferNombre: "Juan Ramon Quintana", ChoferLicencia: "RTT-567",
		VehiculoPlaca: "RFT-87867", VehiculoTipo: "VOLVO",
		ProveedorNombre: "JONI FERNANDES PALLI",
		ProductoNombre: "PLATA", OrigenNombre: "EL ALTO", DestinoNombre: "ORURO - CHARQUITA",
		Operador: "Maria Lopez",
		PesoBruto: 7497, Tara: 4000, PesoNeto: 3497,
		CostoBs: 500, FleteBs: 0,
	}
	cfg := &boleta.Config{
		Salida: "Salida a Potosi", Direccion: "Tarija km 2",
		Celular1: "2782323", Celular2: "1234343",
		Ciudad: "Potosi - Bolivia",
		TituloSuperior: "BALANZA DE PESAJE DIGITAL",
		Marca: "PRIMAVERA",
		TituloDocumento: "BOLETA DE PESAJE",
		Acreditacion: "CALTEC",
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