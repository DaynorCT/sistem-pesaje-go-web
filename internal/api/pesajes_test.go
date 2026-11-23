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

func TestPesajesEscritorio(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	secret := "test-secret"
	refresh := "test-refresh"
	email := config.EmailAdministrador
	_, err = sqlDB.Exec(`UPDATE usuarios SET password_hash = ? WHERE email = ?`, clave.Hash("admin123"), email)
	if err != nil {
		t.Fatal(err)
	}

	authH := &api.AuthHandlers{DB: sqlDB, JWTSecret: secret, RefreshSecret: refresh}
	loginBody, _ := json.Marshal(map[string]string{"usuario": email, "contrasena": "admin123"})
	tokenRec := httptest.NewRecorder()
	authH.Token(tokenRec, httptest.NewRequest(http.MethodPost, "/api/auth/token", bytes.NewReader(loginBody)))
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("token status %d body %s", tokenRec.Code, tokenRec.Body.String())
	}
	var tokenResp map[string]any
	_ = json.Unmarshal(tokenRec.Body.Bytes(), &tokenResp)
	access, _ := tokenResp["access_token"].(string)

	h := &api.PesajesHandler{DB: sqlDB, JWTSecret: secret}
	payload := map[string]any{
		"client_event_id": "evento-test-1",
		"peso_bruto":      30000,
		"tara":            8000,
		"estado_balanza":  "FINALIZADO",
	}
	b, _ := json.Marshal(payload)
	post := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/pesajes", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", "access_token="+access)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	if rec := post(); rec.Code != http.StatusOK {
		t.Fatalf("post1 %d %s", rec.Code, rec.Body.String())
	}
	rec2 := post()
	if rec2.Code != http.StatusOK {
		t.Fatalf("post2 %d %s", rec2.Code, rec2.Body.String())
	}
	var dup map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &dup)
	if dup["duplicado"] != true {
		t.Fatalf("se esperaba duplicado: %v", dup)
	}

	ultimo := &api.UltimoPesoHandler{DB: sqlDB, JWTSecret: secret}
	pBody, _ := json.Marshal(map[string]float64{"peso": 12345})
	pReq := httptest.NewRequest(http.MethodPost, "/api/pesajes/ultimo-peso", bytes.NewReader(pBody))
	pReq.Header.Set("Content-Type", "application/json")
	pReq.Header.Set("Cookie", "access_token="+access)
	pRec := httptest.NewRecorder()
	ultimo.ServeHTTP(pRec, pReq)
	if pRec.Code != http.StatusOK {
		t.Fatalf("ultimo post %d", pRec.Code)
	}
	gRec := httptest.NewRecorder()
	ultimo.ServeHTTP(gRec, httptest.NewRequest(http.MethodGet, "/api/pesajes/ultimo-peso", nil))
	var live map[string]any
	_ = json.Unmarshal(gRec.Body.Bytes(), &live)
	if live["peso"] != float64(12345) {
		t.Fatalf("peso vivo %v", live["peso"])
	}
}
