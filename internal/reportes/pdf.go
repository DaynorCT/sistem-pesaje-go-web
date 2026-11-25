package reportes

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode"

	"sistema-pesaje-web/internal/boleta"

	"github.com/jung-kurt/gofpdf/v2"
)

const (
	pageAnchoMM   = 215.9
	pageAltoMM    = 279.4
	margenLateral = 9.0
	margenInferior = 18.0

	azulDatosCabR = 37
	azulDatosCabG = 94
	azulDatosCabB = 166

	fondoBarraR   = 232
	fondoBarraG   = 236
	fondoBarraB   = 242
	bordeBarraR   = 195
	bordeBarraG   = 202
	bordeBarraB   = 212
	grosorBordeMM = 0.35

	altoFilaTabla = 6.2
	altoCabTabla  = 7.0
)

// Encabezado filtros y empresa para el PDF de reporte.
type Encabezado struct {
	Empresa string
	Desde   string
	Hasta   string
	Placa   string
}

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

// GenerarPDF listado de pesajes en tabla (estilo alineado a la boleta).
func GenerarPDF(items []boleta.Datos, enc Encabezado) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	x0 := margenLateral
	ancho := pageAnchoMM - 2*margenLateral
	y := dibujarEncabezadoReporte(pdf, x0, ancho, 12, enc)
	y += 4

	cols := columnasReporte(ancho)
	y = dibujarCabeceraTabla(pdf, x0, y, cols)

	if len(items) == 0 {
		y = dibujarFilaTabla(pdf, x0, y, cols, []string{"-", "-", "-", "-", "-", "-", "-", "-", "Sin registros"})
	} else {
		for i, fila := range items {
			if y+altoFilaTabla > pageAltoMM-margenInferior {
				pdf.AddPage()
				y = dibujarEncabezadoReporte(pdf, x0, ancho, 10, enc)
				y += 2
				pdf.SetFont("Helvetica", "I", 8)
				pdf.SetTextColor(80, 85, 92)
				pdf.Text(x0, y, sinAcentos("(continuacion)"))
				y += 5
				y = dibujarCabeceraTabla(pdf, x0, y, cols)
			}
			celdas := celdasDesdeDatos(i+1, &fila)
			y = dibujarFilaTabla(pdf, x0, y, cols, celdas)
		}
	}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(azulDatosCabR, azulDatosCabG, azulDatosCabB)
	pdf.Text(x0, pageAltoMM-margenInferior+2, sinAcentos(fmt.Sprintf("Total registros: %d", len(items))))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type colReporte struct {
	titulo string
	ancho  float64
}

func columnasReporte(anchoUtil float64) []colReporte {
	// Fracciones del ancho (suman 1.0)
	fr := []struct {
		t string
		p float64
	}{
		{"N", 0.04},
		{"Guia", 0.14},
		{"Placa", 0.09},
		{"Chofer", 0.18},
		{"Proveedor", 0.16},
		{"Bruto", 0.1},
		{"Tara", 0.09},
		{"Neto", 0.1},
		{"Fecha", 0.1},
	}
	out := make([]colReporte, len(fr))
	for i, c := range fr {
		out[i] = colReporte{titulo: c.t, ancho: anchoUtil * c.p}
	}
	return out
}

func dibujarEncabezadoReporte(pdf *gofpdf.Fpdf, x0, ancho, y float64, enc Encabezado) float64 {
	cx := pageAnchoMM / 2
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(azulDatosCabR, azulDatosCabG, azulDatosCabB)
	tit := "REPORTE DE PESAJE"
	w := pdf.GetStringWidth(tit)
	pdf.Text(cx-w/2, y, tit)
	y += 6
	if enc.Empresa != "" {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(35, 38, 45)
		emp := sinAcentos(strings.ToUpper(enc.Empresa))
		w = pdf.GetStringWidth(emp)
		pdf.Text(cx-w/2, y, emp)
		y += 5
	}
	filtro := textoFiltros(enc)
	if filtro != "" {
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(70, 75, 82)
		filtro = sinAcentos(filtro)
		w = pdf.GetStringWidth(filtro)
		pdf.Text(cx-w/2, y, filtro)
		y += 4.5
	}
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(100, 105, 112)
	gen := sinAcentos("Generado: " + time.Now().Format("02/01/2006 15:04"))
	w = pdf.GetStringWidth(gen)
	pdf.Text(cx-w/2, y, gen)
	y += 4
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.Line(x0, y, x0+ancho, y)
	pdf.SetTextColor(0, 0, 0)
	return y
}

func textoFiltros(enc Encabezado) string {
	var p []string
	if d := strings.TrimSpace(enc.Desde); d != "" {
		p = append(p, "Desde "+d)
	}
	if h := strings.TrimSpace(enc.Hasta); h != "" {
		p = append(p, "Hasta "+h)
	}
	if pl := strings.TrimSpace(enc.Placa); pl != "" {
		p = append(p, "Placa "+strings.ToUpper(pl))
	}
	if len(p) == 0 {
		return "Filtros: todos los pesajes activos"
	}
	return "Filtros: " + strings.Join(p, "  |  ")
}

func dibujarCabeceraTabla(pdf *gofpdf.Fpdf, x0, y float64, cols []colReporte) float64 {
	x := x0
	pdf.SetFillColor(fondoBarraR, fondoBarraG, fondoBarraB)
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	for _, c := range cols {
		pdf.Rect(x, y, c.ancho, altoCabTabla, "FD")
		x += c.ancho
	}
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(azulDatosCabR, azulDatosCabG, azulDatosCabB)
	x = x0
	for _, c := range cols {
		centrarEnCelda(pdf, x, c.ancho, y+4.5, sinAcentos(strings.ToUpper(c.titulo)))
		x += c.ancho
	}
	pdf.SetTextColor(0, 0, 0)
	return y + altoCabTabla
}

func dibujarFilaTabla(pdf *gofpdf.Fpdf, x0, y float64, cols []colReporte, valores []string) float64 {
	x := x0
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	for i, c := range cols {
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(252, 253, 255)
		}
		pdf.Rect(x, y, c.ancho, altoFilaTabla, "FD")
		x += c.ancho
	}
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(0, 0, 0)
	x = x0
	for i, c := range cols {
		val := ""
		if i < len(valores) {
			val = valores[i]
		}
		textoEnCelda(pdf, x, c.ancho, y+3.8, val)
		x += c.ancho
	}
	return y + altoFilaTabla
}

func celdasDesdeDatos(n int, d *boleta.Datos) []string {
	guia := strings.TrimSpace(d.Guia)
	if guia == "" {
		guia = "-"
	}
	chofer := strings.TrimSpace(d.ChoferNombre)
	if chofer == "" {
		chofer = "-"
	}
	prov := strings.TrimSpace(d.ProveedorNombre)
	if prov == "" {
		prov = "-"
	}
	placa := strings.TrimSpace(d.VehiculoPlaca)
	if placa == "" {
		placa = "-"
	}
	fh := strings.TrimSpace(d.Fecha + " " + d.Hora)
	if fh == "" {
		fh = "-"
	}
	return []string{
		fmt.Sprintf("%d", n),
		sinAcentos(strings.ToUpper(guia)),
		sinAcentos(strings.ToUpper(placa)),
		sinAcentos(strings.ToUpper(chofer)),
		sinAcentos(strings.ToUpper(prov)),
		boleta.FormatoPesoKg(d.PesoBruto),
		boleta.FormatoPesoKg(d.Tara),
		boleta.FormatoPesoKg(d.PesoNeto),
		sinAcentos(fh),
	}
}

func centrarEnCelda(pdf *gofpdf.Fpdf, x, w, y float64, s string) {
	tw := pdf.GetStringWidth(s)
	if tw > w-2 {
		s = truncarTexto(pdf, s, w-2)
		tw = pdf.GetStringWidth(s)
	}
	pdf.Text(x+(w-tw)/2, y, s)
}

func textoEnCelda(pdf *gofpdf.Fpdf, x, w, y float64, s string) {
	const pad = 1.2
	s = truncarTexto(pdf, s, w-2*pad)
	pdf.Text(x+pad, y, s)
}

func truncarTexto(pdf *gofpdf.Fpdf, s string, maxW float64) string {
	if pdf.GetStringWidth(s) <= maxW {
		return s
	}
	ell := "..."
	for len(s) > 0 {
		s = strings.TrimSpace(s[:len(s)-1])
		if pdf.GetStringWidth(s+ell) <= maxW {
			return s + ell
		}
	}
	return ell
}