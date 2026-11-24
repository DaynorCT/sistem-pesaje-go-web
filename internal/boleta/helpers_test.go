package boleta

import "testing"

func TestFormatoEntero(t *testing.T) {
	if formatoEntero(1234567) != "1.234.567" {
		t.Fatalf("got %q", formatoEntero(1234567))
	}
}

func TestFormatoPesoKg(t *testing.T) {
	if formatoPesoKg(7497) != "7.497" {
		t.Fatalf("got %q", formatoPesoKg(7497))
	}
	if formatoPesoKg(4000) != "4.000" {
		t.Fatalf("got %q", formatoPesoKg(4000))
	}
	if formatoPesoKg(30000) != "30.000" {
		t.Fatalf("got %q", formatoPesoKg(30000))
	}
}

func TestDecodificarLogoDataURL(t *testing.T) {
  if DecodificarLogoDataURL("") != nil {
    t.Fatal("empty")
  }
}
