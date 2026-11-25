package boleta

import "testing"

func TestFormatoEntero(t *testing.T) {
	if formatoEntero(1234567) != "1.234.567" {
		t.Fatalf("got %q", formatoEntero(1234567))
	}
}

func TestFormatoPesoKg(t *testing.T) {
	if FormatoPesoKg(7497) != "7.497" {
		t.Fatalf("got %q", FormatoPesoKg(7497))
	}
	if FormatoPesoKg(4000) != "4.000" {
		t.Fatalf("got %q", FormatoPesoKg(4000))
	}
	if FormatoPesoKg(30000) != "30.000" {
		t.Fatalf("got %q", FormatoPesoKg(30000))
	}
}

func TestDecodificarLogoDataURL(t *testing.T) {
  if DecodificarLogoDataURL("") != nil {
    t.Fatal("empty")
  }
}
