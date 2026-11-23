package boleta

import "testing"

func TestFormatoEntero(t *testing.T) {
  if formatoEntero(1234567) != "1.234.567" {
    t.Fatalf("got %q", formatoEntero(1234567))
  }
}

func TestDecodificarLogoDataURL(t *testing.T) {
  if DecodificarLogoDataURL("") != nil {
    t.Fatal("empty")
  }
}
