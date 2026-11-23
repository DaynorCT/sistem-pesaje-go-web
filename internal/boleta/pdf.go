package boleta

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"unicode"

	"github.com/jung-kurt/gofpdf/v2"
)

const (
	logoXDerMM         = 266.0 // mismo borde derecho que las líneas horizontales del cuerpo
	logoColDerInicioMM = 155.0 // columna derecha del detalle (no invadir con el logo)
	logoYAcreditEtiquetaMM = 19.0
	logoYAcreditTextoMM    = 22.5
	logoYSuperiorMM        = 25.0
	logoLadoMaxDeseMM  = 70.0
	logoMargenMM       = 5.0
	logoYInferiorEncMM = 57.0
	anchoHojaBoletaMM  = 279.4
	altoHojaBoletaMM   = 215.9
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

// GenerarPDF boleta apaisada carta (paridad pesaje-go reportes.GenerarBoleta).
func GenerarPDF(d *Datos, cfg *Config) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("sin datos de pesaje")
	}
	if cfg == nil {
		cfg = &Config{TituloSuperior: "BOLETA DE PESAJE", Marca: "SISTEMA DE PESAJE"}
	}
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 279.4, Ht: 215.9},
	})
	pdf.AddPage()
	dibujarAcreditacionEncabezado(pdf, cfg)
	dibujarLogoBoleta(pdf, LogoSistemaJPEG())
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Text(15, 20, sinAcentos(strings.ToUpper(cfg.Salida)))
	pdf.SetFont("Helvetica", "B", 15)
	pdf.Text(95, 20, sinAcentos(strings.ToUpper(cfg.TituloSuperior)))
	pdf.SetFont("Helvetica", "", 11)
	pdf.Text(15, 25, sinAcentos(strings.ToUpper(cfg.Direccion)))
	pdf.SetFont("Helvetica", "B", 22)
	pdf.Text(110, 30, sinAcentos(strings.ToUpper(cfg.Marca)))
	pdf.SetFont("Helvetica", "", 11)
	pdf.Text(15, 30, "Cel: "+cfg.Celular1)
	pdf.SetFont("Helvetica", "B", 15)
	pdf.Text(108, 40, sinAcentos(strings.ToUpper(cfg.TituloDocumento)))
	pdf.SetFont("Helvetica", "", 11)
	pdf.Text(22, 34, cfg.Celular2)
	pdf.Text(15, 38, sinAcentos(strings.ToUpper(cfg.Ciudad)))
	pdf.Text(15, 56, "Guia: "+sinAcentos(strings.ToUpper(d.Guia)))

	pdf.SetFont("Helvetica", "B", 13)
	pdf.Text(15, 64, "DATOS DEL VEHICULO:")
	pdf.Text(155, 64, "DATOS DEL PESAJE:")
	pdf.SetFont("Helvetica", "", 12)
	filaBoleta(pdf, 71, "Placa:", d.VehiculoPlaca, "Producto:", d.ProductoNombre)
	filaBoleta(pdf, 78, "Chofer:", d.ChoferNombre, "Costo Bs.:", numero(d.CostoBs)+" Bs")
	filaBoleta(pdf, 85, "Licencia:", d.ChoferLicencia, "Origen:", d.OrigenNombre)
	filaBoleta(pdf, 92, "Proveedor:", d.ProveedorNombre, "Destino:", d.DestinoNombre)
	filaBoleta(pdf, 99, "Tipo Vehiculo:", d.VehiculoTipo, "Flete Bs.:", numero(d.FleteBs)+" Bs")

	pdf.SetFont("Helvetica", "", 11)
	fh := fmt.Sprintf("Fecha/Hora (Pes): %s %s", d.Fecha, d.Hora)
	pdf.Text(15, 107, sinAcentos(fh))
	pdf.Text(155, 107, sinAcentos("Fecha/Hora (Imp): "+d.Fecha+" "+d.Hora))

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Text(15, 116, "PESO BRUTO:")
	pdf.Text(115, 116, "PESO TARA:")
	pdf.Text(205, 116, "PESO NETO:")
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Text(15, 122, numero(d.PesoBruto)+" kg")
	pdf.Text(115, 122, numero(d.Tara)+" kg")
	pdf.Text(205, 122, numero(d.PesoNeto)+" kg")

	pdf.SetFont("Helvetica", "", 10)
	pdf.Text(60, 160, "Chofer o Productor")
	pdf.Text(187, 160, "Operador de balanza")

	pdf.Line(15, 59, 266, 59)
	pdf.Line(15, 102, 266, 102)
	pdf.Line(15, 111, 266, 111)
	pdf.Line(15, 126, 266, 126)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func dibujarAcreditacionEncabezado(pdf *gofpdf.Fpdf, cfg *Config) {
	if cfg == nil {
		return
	}
	pdf.SetFont("Helvetica", "B", 8)
	etiqueta := "ACREDITADO POR:"
	textoDerecha(pdf, logoXDerMM, logoYAcreditEtiquetaMM, sinAcentos(etiqueta))
	if strings.TrimSpace(cfg.Acreditacion) != "" {
		textoDerecha(pdf, logoXDerMM, logoYAcreditTextoMM, sinAcentos(cfg.Acreditacion))
	}
}

func textoDerecha(pdf *gofpdf.Fpdf, bordeDerXMm, y float64, s string) {
	w := pdf.GetStringWidth(s)
	x := bordeDerXMm - w
	if x < logoColDerInicioMM {
		x = logoColDerInicioMM
	}
	pdf.Text(x, y, s)
}

func dibujarLogoBoleta(pdf *gofpdf.Fpdf, logoJPEG []byte) {
	if len(logoJPEG) == 0 {
		return
	}
	wPx, hPx, imgType := dimensionesImagen(logoJPEG)
	if imgType == "" || wPx <= 0 || hPx <= 0 {
		return
	}
	wMm, hMm := tamanoLogoEnHoja(wPx, hPx)
	name := "logo-boleta"
	opts := gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	_ = pdf.RegisterImageOptionsReader(name, opts, bytes.NewReader(logoJPEG))
	xLogo := logoXDerMM - wMm
	if xLogo < logoColDerInicioMM {
		xLogo = logoColDerInicioMM
	}
	pdf.ImageOptions(name, xLogo, logoYSuperiorMM, wMm, hMm, false, opts, 0, "")
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

func tamanoLogoEnHoja(pxAncho, pxAlto int) (wMm, hMm float64) {
	cupoAncho := logoLadoMaxDeseMM
	cupoAnchoMax := logoXDerMM - logoColDerInicioMM - logoMargenMM
	if cupoAncho > cupoAnchoMax {
		cupoAncho = cupoAnchoMax
	}
	cupoAltoEnc := logoYInferiorEncMM - logoYSuperiorMM - logoMargenMM
	cupoAltoHoja := altoHojaBoletaMM - logoYSuperiorMM - logoMargenMM
	cupoAlto := logoLadoMaxDeseMM
	if cupoAlto > cupoAltoEnc {
		cupoAlto = cupoAltoEnc
	}
	if cupoAlto > cupoAltoHoja {
		cupoAlto = cupoAltoHoja
	}
	if cupoAncho < 1 {
		cupoAncho = 1
	}
	if cupoAlto < 1 {
		cupoAlto = 1
	}
	proporcion := float64(pxAncho) / float64(pxAlto)
	if float64(pxAncho)/cupoAncho > float64(pxAlto)/cupoAlto {
		wMm = cupoAncho
		hMm = cupoAncho / proporcion
	} else {
		hMm = cupoAlto
		wMm = cupoAlto * proporcion
	}
	if hMm > cupoAlto {
		hMm = cupoAlto
		wMm = cupoAlto * proporcion
	}
	if wMm > cupoAncho {
		wMm = cupoAncho
		hMm = cupoAncho / proporcion
	}
	return wMm, hMm
}

func filaBoleta(pdf *gofpdf.Fpdf, y float64, ei, vi, ed, vd string) {
	pdf.Text(15, y, sinAcentos(ei))
	pdf.Text(80, y, sinAcentos(strings.ToUpper(vi)))
	pdf.Text(155, y, sinAcentos(ed))
	pdf.Text(220, y, sinAcentos(strings.ToUpper(vd)))
}
