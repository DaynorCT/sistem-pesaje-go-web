package boleta

import (
	"fmt"
	"strings"
)

func formatoEntero(n int) string {
	if n < 0 {
		return "-" + formatoEntero(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var out strings.Builder
	primera := len(s) % 3
	if primera == 0 {
		primera = 3
	}
	out.WriteString(s[:primera])
	for i := primera; i < len(s); i += 3 {
		out.WriteByte('.')
		out.WriteString(s[i : i+3])
	}
	return out.String()
}

func formatoBs(n int) string {
	return fmt.Sprintf("%s.00", formatoEntero(n))
}

// FormatoPesoKg: bajo 10 000 suele ser kg×1000 (7.497); arriba, entero en kg (30.000).
func FormatoPesoKg(v int) string {
	if v < 0 {
		return "-" + FormatoPesoKg(-v)
	}
	if v == 0 {
		return "0.000"
	}
	if v < 10000 {
		return fmt.Sprintf("%.3f", float64(v)/1000.0)
	}
	return formatoEntero(v)
}