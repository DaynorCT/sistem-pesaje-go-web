package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"sistema-pesaje-web/internal/api"
	"sistema-pesaje-web/internal/clave"
	"sistema-pesaje-web/internal/config"
	"sistema-pesaje-web/internal/db"
)

func TestUsuariosCRUD(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	secret := "test-secret"
	_, _ = sqlDB.Exec(`UPDATE usuarios SET password_hash = ? WHERE email = ?`, clave.Hash("admin123"), config.EmailAdministrador)

	authH := &api.AuthHandlers{DB: sqlDB, JWTSecret: secret, RefreshSecret: "r"}
	loginBody, _ := json.Marshal(map[string]string{"usuario": config.EmailAdministrador, "contrasena": "admin123"})
	rec := httptest.NewRecorder()
	authH.Token(rec, httptest.NewRequest(http.MethodPost, "/api/auth/token", bytes.NewReader(loginBody)))
	var tok map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &tok)
	access, _ := tok["access_token"].(string)

	h := &api.UsuariosHandler{DB: sqlDB, JWTSecret: secret}

	gRec := httptest.NewRecorder()
	h.ServeHTTP(gRec, httptest.NewRequest(http.MethodGet, "/api/usuarios", nil))
	if gRec.Code != http.StatusOK {
		t.Fatalf("list %d", gRec.Code)
	}

	createBody, _ := json.Marshal(map[string]any{
		"nombre": "Juan", "apellido_paterno": "Pérez", "correo": "juan@test.com", "clave": "secret123", "rol": "usuario",
	})
	cReq := httptest.NewRequest(http.MethodPost, "/api/usuarios", bytes.NewReader(createBody))
	cReq.Header.Set("Content-Type", "application/json")
	cReq.Header.Set("Cookie", "access_token="+access)
	cRec := httptest.NewRecorder()
	h.ServeHTTP(cRec, cReq)
	if cRec.Code != http.StatusOK {
		t.Fatalf("create %d %s", cRec.Code, cRec.Body.String())
	}
	var cResp map[string]any
	_ = json.Unmarshal(cRec.Body.Bytes(), &cResp)
	id := int(cResp["id"].(float64))

	upBody, _ := json.Marshal(map[string]any{
		"id": id, "nombre": "Juan", "correo": "juan@test.com", "rol": "administrador", "estado": "ACTIVO",
	})
	uReq := httptest.NewRequest(http.MethodPut, "/api/usuarios", bytes.NewReader(upBody))
	uReq.Header.Set("Content-Type", "application/json")
	uReq.Header.Set("Cookie", "access_token="+access)
	uRec := httptest.NewRecorder()
	h.ServeHTTP(uRec, uReq)
	if uRec.Code != http.StatusOK {
		t.Fatalf("update %d %s", uRec.Code, uRec.Body.String())
	}
}