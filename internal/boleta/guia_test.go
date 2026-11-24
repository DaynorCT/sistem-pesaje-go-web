package boleta

import (
	"strings"
	"testing"
)

func TestGuiaTexto(t *testing.T) {
	d := &Datos{ID: 1, Guia: "PESO-2026-000001", Fecha: "23/11/2026"}
	got := guiaTexto(d)
	if !strings.Contains(got, "PESO-2026-000001") {
		t.Fatalf("guia: %q", got)
	}
	if !strings.HasPrefix(got, "GUIA N") {
		t.Fatalf("prefijo: %q", got)
	}
	d2 := &Datos{ID: 42, Fecha: "23/11/2026"}
	got2 := guiaTexto(d2)
	if !strings.Contains(got2, "PESO-2026-000042") {
		t.Fatalf("fallback: %q", got2)
	}
}