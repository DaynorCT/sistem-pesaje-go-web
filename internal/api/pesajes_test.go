package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
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

	// Pesaje EN_PROCESO para finalizar desde la web
	procBody, _ := json.Marshal(map[string]any{
		"peso_bruto":     25000,
		"tara":           8000,
		"estado_balanza": "EN_PROCESO",
	})
	procReq := httptest.NewRequest(http.MethodPost, "/api/pesajes", bytes.NewReader(procBody))
	procReq.Header.Set("Content-Type", "application/json")
	procReq.Header.Set("Cookie", "access_token="+access)
	procRec := httptest.NewRecorder()
	h.ServeHTTP(procRec, procReq)
	if procRec.Code != http.StatusOK {
		t.Fatalf("pesaje proceso %d %s", procRec.Code, procRec.Body.String())
	}
	var procResp map[string]any
	_ = json.Unmarshal(procRec.Body.Bytes(), &procResp)
	id := int64(procResp["id"].(float64))

	finReq := httptest.NewRequest(http.MethodPost, "/api/pesajes/"+strconv.FormatInt(id, 10)+"/finalizar", nil)
	finReq.Header.Set("Cookie", "access_token="+access)
	finReq.SetPathValue("id", strconv.FormatInt(id, 10))
	finRec := httptest.NewRecorder()
	h.FinalizarByID(finRec, finReq)
	if finRec.Code != http.StatusOK {
		t.Fatalf("finalizar %d %s", finRec.Code, finRec.Body.String())
	}
}
