package reportes

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"sistema-pesaje-web/internal/boleta"

	"github.com/jung-kurt/gofpdf/v2"
)

func sinAcentos(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'Á':
			b.WriteRune('a')
		case 'é', 'É':
			b.WriteRune('e')
		case 'í', 'Í':
			b.WriteRune('i')
		case 'ó', 'Ó':
			b.WriteRune('o')
		case 'ú', 'Ú', 'ü', 'Ü':
			b.WriteRune('u')
		case 'ñ', 'Ñ':
			b.WriteRune('n')
		default:
			if r > unicode.MaxASCII && r != ' ' {
				b.WriteRune('?')
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func numero(n int) string {
	return fmt.Sprintf("%d", n)
}

// GenerarPDF vertical (paridad pesaje-go reportes.GenerarReporte).
func GenerarPDF(items []boleta.Datos, nombreEmpresa string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Text(75, 25, "REPORTE DE PESAJE")
	if nombreEmpresa != "" {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.Text(60, 32, sinAcentos(strings.ToUpper(nombreEmpresa)))
	}
	y := 45.0
	for i, fila := range items {
		if y > 240 {
			pdf.AddPage()
			y = 25
		}
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Text(15, y, fmt.Sprintf("Nro. %d", i+1))
		y += 5
		pdf.SetFont("Helvetica", "", 10)
		par(pdf, y, "GUIA:", fila.Guia, "COSTO", "Bs "+numero(fila.CostoBs))
		y += 5
		par(pdf, y, "CHOFER", fila.ChoferNombre, "FLETE", "Bs "+numero(fila.FleteBs))
		y += 5
		par(pdf, y, "VEHICULO", fila.VehiculoTipo+" - "+fila.VehiculoPlaca, "", "")
		y += 5
		par(pdf, y, "NRO. DOC.", fila.ChoferCI, "", "")
		y += 5
		par(pdf, y, "LICENCIA", fila.ChoferLicencia, "", "")
		y += 5
		par(pdf, y, "PROVEEDOR", fila.ProveedorNombre, "", "")
		y += 5
		par(pdf, y, "PRODUCTO", fila.ProductoNombre, "", "")
		y += 5
		par(pdf, y, "ORIGEN", fila.OrigenNombre, "", "")
		y += 5
		par(pdf, y, "DESTINO", fila.DestinoNombre, "", "")
		y += 5
		par(pdf, y, "P. BRUTO", numero(fila.PesoBruto)+" kg", "", "")
		y += 5
		par(pdf, y, "TARA", numero(fila.Tara)+" kg", "", "")
		y += 5
		par(pdf, y, "PESO NETO", numero(fila.PesoNeto)+" kg", "FECHA/HORA", fila.Fecha+" "+fila.Hora)
		y += 10
	}
	pdf.SetFont("Helvetica", "B", 11)
	pdf.Text(15, y+6, fmt.Sprintf("Total registros: %d", len(items)))
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func par(pdf *gofpdf.Fpdf, y float64, ei, vi, ed, vd string) {
	pdf.Text(15, y, sinAcentos(ei))
	pdf.Text(45, y, sinAcentos(vi))
	if ed != "" {
		pdf.Text(110, y, sinAcentos(ed))
		pdf.Text(145, y, sinAcentos(vd))
	}
}
