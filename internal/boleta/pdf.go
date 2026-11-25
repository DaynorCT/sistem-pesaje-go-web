package boleta

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jung-kurt/gofpdf/v2"
)

const (
	pageAnchoMM   = 215.9
	pageAltoMM    = 279.4
	margenLateral = 9.0
	altoMitadMM   = pageAltoMM / 2

	azulR = 37
	azulG = 94
	azulB = 166

	azulTituloR = 37
	azulTituloG = 94
	azulTituloB = 166

	azulDatosCabR = 37
	azulDatosCabG = 94
	azulDatosCabB = 166

	azulSuaveR = 210
	azulSuaveG = 224
	azulSuaveB = 242

	fondoBarraR   = 232
	fondoBarraG   = 236
	fondoBarraB   = 242
	bordeBarraR   = 195
	bordeBarraG   = 202
	bordeBarraB   = 212
	grosorBordeMM = 0.35

	fondoGuiaR = 233
	fondoGuiaG = 239
	fondoGuiaB = 245
)

// GenerarPDF boleta vertical carta: 2 comprobantes por hoja (cliente + balanza).
func GenerarPDF(d *Datos, cfg *Config, logo []byte) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("sin datos de pesaje")
	}
	if cfg == nil {
		cfg = &Config{TituloSuperior: "BOLETA DE PESAJE", Marca: "SISTEMA DE PESAJE"}
	}
	if len(logo) == 0 {
		logo = LogoSistemaJPEG()
	}

	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	ahora := time.Now()
	fechaImp := ahora.Format("02/01/2006")
	horaImp := ahora.Format("15:04")

	dibujarBoletaMitad(pdf, d, cfg, logo, 0, 0, fechaImp, horaImp)
	lineaCorteMitad(pdf)
	dibujarBoletaMitad(pdf, d, cfg, logo, altoMitadMM, 1, fechaImp, horaImp)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func lineaCorteMitad(pdf *gofpdf.Fpdf) {
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.SetDashPattern([]float64{1.5, 1.5}, 0)
	y := altoMitadMM
	pdf.Line(margenLateral, y, pageAnchoMM-margenLateral, y)
	pdf.SetDashPattern(nil, 0)
}

func dibujarBoletaMitad(pdf *gofpdf.Fpdf, d *Datos, cfg *Config, logo []byte, y0 float64, copiaIdx int, fechaImp, horaImp string) {
	x0 := margenLateral
	ancho := pageAnchoMM - 2*margenLateral

	// --- Encabezado: izquierda | centro | derecha ---
	yEnc := y0 + 7
	const pasoEncIzq = 4.5
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(45, 45, 48)
	pdf.Text(x0, yEnc, sinAcentos(strings.ToUpper(cfg.Salida)))
	pdf.SetFont("Helvetica", "", 9)
	pdf.Text(x0, yEnc+pasoEncIzq, sinAcentos(strings.ToUpper(cfg.Direccion)))
	cel := "Cel: " + cfg.Celular1
	if strings.TrimSpace(cfg.Celular2) != "" {
		cel += " - " + cfg.Celular2
	}
	pdf.Text(x0, yEnc+2*pasoEncIzq, cel)
	pdf.Text(x0, yEnc+3*pasoEncIzq, sinAcentos(strings.ToUpper(cfg.Ciudad)))

	cx := pageAnchoMM / 2
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(azulDatosCabR, azulDatosCabG, azulDatosCabB)
	titulo := sinAcentos(strings.ToUpper(cfg.TituloSuperior))
	centrarTexto(pdf, cx, yEnc, titulo)
	pdf.SetFont("Helvetica", "B", 15)
	centrarTextoEspaciado(pdf, cx, yEnc+7.5, sinAcentos(cfg.Marca), 1.2)
	pdf.SetFont("Helvetica", "B", 9.5)
	pdf.SetTextColor(25, 25, 28)
	doc := sinAcentos(strings.ToUpper(cfg.TituloDocumento))
	centrarTexto(pdf, cx, yEnc+14, doc)
	pdf.SetTextColor(0, 0, 0)

	dibujarLogoBoletaVertical(pdf, logo, cfg, y0+5, copiaIdx)

	const (
		altoBar    = 9.0
		fontGuiaPt = 9.5
	)
	yBar := y0 + 25
	prefGuia, numGuia := guiaPartes(d)
	pdf.SetFont("Helvetica", "B", 9)
	anchoFondoGuia := pdf.GetStringWidth(prefGuia+numGuia) + 6
	if anchoFondoGuia > ancho {
		anchoFondoGuia = ancho
	}

	pdf.SetFillColor(fondoGuiaR, fondoGuiaG, fondoGuiaB)
	pdf.Rect(x0, yBar, anchoFondoGuia, altoBar, "F")
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.Line(x0, yBar, x0+ancho, yBar)
	pdf.Line(x0, yBar+altoBar, x0+ancho, yBar+altoBar)

	dibujarGuiaEnBarraIzq(pdf, x0+3, yBar+5.6, fontGuiaPt, prefGuia, numGuia)
	pdf.SetTextColor(0, 0, 0)

	// Dos columnas lado a lado (como la imagen de referencia)
	const (
		gapEntreBloques = 4.0
		gapTrasFecha    = 3.5
	)
	gap := 3.5
	colW := (ancho - gap) / 2
	altoCol := 42.0
	yCols := yBar + altoBar + gapEntreBloques
	dibujarCajaDatosColumna(pdf, x0, yCols, colW, altoCol, "DATOS DEL VEHICULO", [][2]string{
		{"Placa", d.VehiculoPlaca},
		{"Chofer", d.ChoferNombre},
		{"Licencia", d.ChoferLicencia},
		{"Proveedor", d.ProveedorNombre},
		{"Tipo Vehiculo", d.VehiculoTipo},
	})
	dibujarCajaDatosColumna(pdf, x0+colW+gap, yCols, colW, altoCol, "DATOS DEL PESAJE", [][2]string{
		{"Producto", d.ProductoNombre},
		{"Costo Bs.", formatoBs(d.CostoBs)},
		{"Origen", d.OrigenNombre},
		{"Destino", d.DestinoNombre},
		{"Flete Bs.", formatoBs(d.FleteBs)},
	})

	// Fecha / hora: una sola caja (como antes); Imp alineado con DATOS DEL PESAJE
	yFh := yCols + altoCol + gapEntreBloques
	xColDer := x0 + colW + gap
	const altoFh = 6.5
	cajaFecha(pdf, x0, yFh, ancho, altoFh)
	xSepFh := x0 + colW + gap/2
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.Line(xSepFh, yFh+1.2, xSepFh, yFh+altoFh-1.2)
	pdf.SetTextColor(60, 65, 72)
	textoFechaHoraBoleta(pdf, x0+3, yFh+4.3, "Fecha/Hora (Pes)", d.Fecha, d.Hora)
	textoFechaHoraBoleta(pdf, xColDer+3, yFh+4.3, "Fecha/Hora (Imp)", fechaImp, horaImp)
	pdf.SetTextColor(0, 0, 0)

	// Tres pesos
	yPesos := yFh + altoFh + gapTrasFecha
	altoPeso := 22.0
	gapP := 3.5
	triW := (ancho - 2*gapP) / 3
	dibujarCajaPeso(pdf, x0, yPesos, triW, altoPeso, "PESO BRUTO", FormatoPesoKg(d.PesoBruto)+" kg", false)
	dibujarCajaPeso(pdf, x0+triW+gapP, yPesos, triW, altoPeso, "PESO TARA", FormatoPesoKg(d.Tara)+" kg", false)
	dibujarCajaPeso(pdf, x0+2*(triW+gapP), yPesos, triW, altoPeso, "PESO NETO", FormatoPesoKg(d.PesoNeto)+" kg", true)

	const gapAntesFirma = 14.0
	yFirma := yPesos + altoPeso + gapAntesFirma
	mitad := (ancho - 10) / 2
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.Line(x0, yFirma, x0+mitad, yFirma)
	pdf.Line(x0+mitad+10, yFirma, x0+ancho, yFirma)
	pdf.SetFont("Helvetica", "", 9.0)
	pdf.SetTextColor(90, 95, 102)
	centrarTextoEnRango(pdf, x0, x0+mitad, yFirma+4, "Chofer o Productor")
	centrarTextoEnRango(pdf, x0+mitad+10, x0+ancho, yFirma+4, "Operador de balanza")
	pdf.SetTextColor(0, 0, 0)
}

// prefijoGuiaPDF = "GUIA N°:" (N + grado WinAnsi) para Helvetica core.
const prefijoGuiaPDF = "GUIA N\xb0: "

func guiaPartes(d *Datos) (prefijo, numero string) {
	prefijo = prefijoGuiaPDF
	g := strings.TrimSpace(d.Guia)
	if g != "" {
		return prefijo, sinAcentos(strings.ToUpper(g))
	}
	year := time.Now().Year()
	if p := strings.Split(strings.TrimSpace(d.Fecha), "/"); len(p) == 3 {
		if y, err := strconv.Atoi(p[2]); err == nil && y > 1900 {
			year = y
		}
	}
	return prefijo, fmt.Sprintf("PESO-%d-%06d", year, d.ID)
}

func guiaTexto(d *Datos) string {
	p, n := guiaPartes(d)
	return p + n
}

func centrarTextoEspaciado(pdf *gofpdf.Fpdf, cx, y float64, s string, extraMM float64) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return
	}
	pdf.SetTextColor(azulR, azulG, azulB)
	var total float64
	for _, r := range s {
		if r == ' ' {
			continue
		}
		total += pdf.GetStringWidth(string(r)) + extraMM
	}
	if total > extraMM {
		total -= extraMM
	}
	x := cx - total/2
	for _, r := range s {
		if r == ' ' {
			continue
		}
		ch := string(r)
		pdf.Text(x, y, sinAcentos(ch))
		x += pdf.GetStringWidth(ch) + extraMM
	}
}

func dibujarGuiaEnBarraIzq(pdf *gofpdf.Fpdf, x, y, tamFuente float64, prefijo, numero string) {
	pdf.SetFont("Helvetica", "B", tamFuente)
	wPref := pdf.GetStringWidth(prefijo)
	pdf.SetTextColor(azulDatosCabR, azulDatosCabG, azulDatosCabB)
	pdf.Text(x, y, prefijo)
	pdf.SetTextColor(25, 28, 35)
	pdf.Text(x+wPref, y, numero)
}

func centrarTexto(pdf *gofpdf.Fpdf, cx, y float64, s string) {
	w := pdf.GetStringWidth(s)
	pdf.Text(cx-w/2, y, s)
}

func centrarTextoEnRango(pdf *gofpdf.Fpdf, x1, x2, y float64, s string) {
	w := pdf.GetStringWidth(s)
	cx := (x1 + x2) / 2
	pdf.Text(cx-w/2, y, s)
}

func textoFechaHoraBoleta(pdf *gofpdf.Fpdf, x, y float64, etiqueta, fecha, hora string) {
	const tam = 9.0
	etiq := sinAcentos(etiqueta + ": ")
	val := sinAcentos(strings.TrimSpace(fecha + " " + hora))
	pdf.SetFont("Helvetica", "", tam)
	pdf.Text(x, y, etiq)
	pdf.SetFont("Helvetica", "B", tam)
	pdf.Text(x+pdf.GetStringWidth(etiq), y, val)
}

func cajaFecha(pdf *gofpdf.Fpdf, x, y, w, h float64) {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.RoundedRect(x, y, w, h, 1.5, "1234", "FD")
}

func dibujarCajaDatosColumna(pdf *gofpdf.Fpdf, x, y, w, h float64, titulo string, filas [][2]string) {
	const altoCab = 6.5
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(x, y, w, h, 2, "1234", "FD")
	pdf.SetFillColor(fondoBarraR, fondoBarraG, fondoBarraB)
	pdf.RoundedRect(x, y, w, altoCab, 2, "12", "F")
	pdf.Rect(x, y+altoCab-0.8, w, 0.8, "F")
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(azulDatosCabR, azulDatosCabG, azulDatosCabB)
	tit := sinAcentos(titulo)
	pdf.Text(x+2.5, y+4.4, tit)

	const (
		etiqX    = 2.5
		valorX   = 26.0
		fontEtiq = 9.0
		fontVal  = 9.5
	)
	bodyTop := y + altoCab + 2.5
	bodyH := h - altoCab - 4
	paso := bodyH / float64(len(filas))
	const centradoFila = 0.58
	for i, par := range filas {
		if i > 0 {
			sepY := bodyTop + paso*float64(i)
			pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
			pdf.SetLineWidth(grosorBordeMM)
			pdf.Line(x+2, sepY, x+w-2, sepY)
		}
		ry := bodyTop + paso*float64(i) + paso*centradoFila
		etiq := sinAcentos(par[0] + ":")
		val := textoValorBoleta(par[1])

		pdf.SetFont("Helvetica", "", fontEtiq)
		pdf.SetTextColor(0, 0, 0)
		pdf.Text(x+etiqX, ry, etiq)

		tamVal := fontVal
		pdf.SetFont("Helvetica", "B", tamVal)
		maxW := w - valorX - 2.5
		for tamVal >= 5.5 && pdf.GetStringWidth(val) > maxW {
			tamVal -= 0.5
			pdf.SetFont("Helvetica", "B", tamVal)
		}
		pdf.SetTextColor(0, 0, 0)
		pdf.Text(x+valorX, ry, val)
	}
	pdf.SetLineWidth(grosorBordeMM)
	pdf.SetTextColor(0, 0, 0)
}

func textoValorBoleta(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if esValorNumericoBoleta(s) {
		return sinAcentos(s)
	}
	return sinAcentos(strings.ToUpper(s))
}

func esValorNumericoBoleta(s string) bool {
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '.' || r == ',':
		case i == 0 && r == '-':
		default:
			return false
		}
	}
	return true
}

func dibujarCajaPeso(pdf *gofpdf.Fpdf, x, y, w, h float64, titulo, valor string, _ bool) {
	pdf.SetDrawColor(bordeBarraR, bordeBarraG, bordeBarraB)
	pdf.SetLineWidth(grosorBordeMM)
	pdf.SetFillColor(250, 251, 253)
	pdf.RoundedRect(x, y, w, h, 2.5, "1234", "FD")
	const (
		fontTitPeso = 9.5
		fontValPeso = 15.0
	)
	pdf.SetFont("Helvetica", "", fontTitPeso)
	pdf.SetTextColor(0, 0, 0)
	centrarTextoEnRango(pdf, x, x+w, y+5.5, sinAcentos(titulo))
	tamVal := fontValPeso
	pdf.SetFont("Helvetica", "B", tamVal)
	maxW := w - 4
	for tamVal >= 9 && pdf.GetStringWidth(valor) > maxW {
		tamVal -= 0.5
		pdf.SetFont("Helvetica", "B", tamVal)
	}
	pdf.SetTextColor(0, 0, 0)
	centrarTextoEnRango(pdf, x, x+w, y+h*0.62, valor)
	pdf.SetTextColor(0, 0, 0)
}

func dibujarLogoBoletaVertical(pdf *gofpdf.Fpdf, logo []byte, cfg *Config, y0 float64, copiaIdx int) {
	xDer := pageAnchoMM - margenLateral
	pdf.SetFont("Helvetica", "B", 6)
	pdf.SetTextColor(100, 105, 112)
	etiq := sinAcentos("ACREDITADO POR:")
	wE := pdf.GetStringWidth(etiq)
	pdf.Text(xDer-wE, y0, etiq)

	if len(logo) == 0 {
		if acr := strings.TrimSpace(cfg.Acreditacion); acr != "" {
			pdf.SetFont("Helvetica", "B", 7)
			wa := pdf.GetStringWidth(acr)
			pdf.Text(xDer-wa, y0+5, sinAcentos(strings.ToUpper(acr)))
		}
		return
	}
	wPx, hPx, imgType := dimensionesImagen(logo)
	if imgType == "" || wPx <= 0 || hPx <= 0 {
		return
	}
	const maxW, maxH = 42.0, 20.0
	wMm, hMm := encajarImagen(wPx, hPx, maxW, maxH)
	xLogo := xDer - wMm
	colMinX := pageAnchoMM * 0.58
	if xLogo < colMinX {
		xLogo = colMinX
	}
	name := fmt.Sprintf("logo-boleta-%d", copiaIdx)
	opts := gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	_ = pdf.RegisterImageOptionsReader(name, opts, bytes.NewReader(logo))
	pdf.ImageOptions(name, xLogo, y0+4, wMm, hMm, false, opts, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func encajarImagen(pxAncho, pxAlto int, maxW, maxH float64) (wMm, hMm float64) {
	proporcion := float64(pxAncho) / float64(pxAlto)
	if maxW/proporcion <= maxH {
		wMm = maxW
		hMm = maxW / proporcion
	} else {
		hMm = maxH
		wMm = maxH * proporcion
	}
	return wMm, hMm
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
		case 'º':
			b.WriteString("o")
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

func dimensionesImagen(b []byte) (w, h int, imgType string) {
	cfg, formato, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return 0, 0, ""
	}
	switch formato {
	case "jpeg":
		imgType = "JPG"
	case "png":
		imgType = "PNG"
	default:
		return 0, 0, ""
	}
	return cfg.Width, cfg.Height, imgType
}
