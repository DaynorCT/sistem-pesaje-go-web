package boleta

import (
	"context"
	"database/sql"
	"encoding/base64"
	"strings"
)

func ObtenerLogoEmpresa(ctx context.Context, conn *sql.DB) []byte {
	var logo sql.NullString
	err := conn.QueryRowContext(ctx, `
SELECT logo FROM empresas WHERE estado = 'ACTIVO' ORDER BY id DESC LIMIT 1`).Scan(&logo)
	if err != nil || !logo.Valid || strings.TrimSpace(logo.String) == "" {
		return nil
	}
	return DecodificarLogoDataURL(logo.String)
}

func DecodificarLogoDataURL(s string) []byte {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	payload := s
	if i := strings.Index(s, ","); i >= 0 && strings.HasPrefix(strings.ToLower(s), "data:") {
		payload = s[i+1:]
	}
	b, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		b, err = base64.RawStdEncoding.DecodeString(payload)
		if err != nil {
			return nil
		}
	}
	if len(b) == 0 {
		return nil
	}
	return b
}

func ResolverLogoParaPDF(ctx context.Context, conn *sql.DB) []byte {
	if conn != nil {
		if b := ObtenerLogoEmpresa(ctx, conn); len(b) > 0 {
			return b
		}
	}
	return LogoSistemaJPEG()
}
