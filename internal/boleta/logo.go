package boleta

import _ "embed"

//go:embed assets/logo-pesaje.jpg
var logoPesajeJPEG []byte

// LogoSistemaJPEG logo embebido para el PDF (paridad pesaje-go).
func LogoSistemaJPEG() []byte {
	if len(logoPesajeJPEG) == 0 {
		return nil
	}
	out := make([]byte, len(logoPesajeJPEG))
	copy(out, logoPesajeJPEG)
	return out
}
